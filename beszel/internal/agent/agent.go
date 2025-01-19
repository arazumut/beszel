// Paket agent, ajan'ın SSH sunucusunu ve sistem istatistikleri toplamasını yönetir.
package agent

import (
	"beszel"
	"beszel/internal/entities/system"
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v4/common"
)

type Agent struct {
	debug            bool                       // LOG_LEVEL debug olarak ayarlandığında true
	zfs              bool                       // Sistem arcstats'e sahip olduğunda true
	memCalc          string                     // Bellek hesaplama formülü
	fsNames          []string                   // İzlenen dosya sistemi cihaz adlarının listesi
	fsStats          map[string]*system.FsStats // Her dosya sistemi için disk istatistiklerini takip eder
	netInterfaces    map[string]struct{}        // Tüm geçerli ağ arayüzlerini saklar
	netIoStats       system.NetIoStats          // Bant genişliği kullanımını takip eder
	dockerManager    *dockerManager             // Docker API isteklerini yönetir
	sensorsContext   context.Context            // Sensörler için sys konumunu geçersiz kılmak için sensörler bağlamı
	sensorsWhitelist map[string]struct{}        // İzlenecek sensörlerin listesi
	systemInfo       system.Info                // Ana sistem bilgisi
	gpuManager       *GPUManager                // GPU verilerini yönetir
}

func NewAgent() *Agent {
	return &Agent{
		sensorsContext: context.Background(),
		memCalc:        os.Getenv("MEM_CALC"),
		fsStats:        make(map[string]*system.FsStats),
	}
}

func (a *Agent) Run(pubKey []byte, addr string) {
	// LOG_LEVEL ortam değişkeni tarafından belirlenen bir günlük seviyesi ile slog'u ayarlayın
	if logLevelStr, exists := os.LookupEnv("LOG_LEVEL"); exists {
		switch strings.ToLower(logLevelStr) {
		case "debug":
			a.debug = true
			slog.SetLogLoggerLevel(slog.LevelDebug)
		case "warn":
			slog.SetLogLoggerLevel(slog.LevelWarn)
		case "error":
			slog.SetLogLoggerLevel(slog.LevelError)
		}
	}

	slog.Debug(beszel.Version)

	// Sensörler bağlamını ayarlayın (sensörler için sys konumunu geçersiz kılmaya izin verir)
	if sysSensors, exists := os.LookupEnv("SYS_SENSORS"); exists {
		slog.Info("SYS_SENSORS", "path", sysSensors)
		a.sensorsContext = context.WithValue(a.sensorsContext,
			common.EnvKey, common.EnvMap{common.HostSysEnvKey: sysSensors},
		)
	}

	// Sensörler beyaz listesini ayarlayın
	if sensors, exists := os.LookupEnv("SENSORS"); exists {
		a.sensorsWhitelist = make(map[string]struct{})
		for _, sensor := range strings.Split(sensors, ",") {
			if sensor != "" {
				a.sensorsWhitelist[sensor] = struct{}{}
			}
		}
	}

	// Sistem bilgilerini / docker yöneticisini başlatın
	a.initializeSystemInfo()
	a.initializeDiskInfo()
	a.initializeNetIoStats()
	a.dockerManager = newDockerManager(a)

	// GPU yöneticisini başlatın
	if gm, err := NewGPUManager(); err != nil {
		slog.Debug("GPU", "err", err)
	} else {
		a.gpuManager = gm
	}

	// Eğer debug modundaysa, istatistikleri yazdırın
	if a.debug {
		slog.Debug("İstatistikler", "data", a.gatherStats())
	}

	a.startServer(pubKey, addr)
}

func (a *Agent) gatherStats() system.CombinedData {
	slog.Debug("İstatistikler alınıyor")
	systemData := system.CombinedData{
		Stats: a.getSystemStats(),
		Info:  a.systemInfo,
	}
	slog.Debug("Sistem istatistikleri", "data", systemData)
	// Docker istatistiklerini ekleyin
	if containerStats, err := a.dockerManager.getDockerStats(); err == nil {
		systemData.Containers = containerStats
		slog.Debug("Docker istatistikleri", "data", systemData.Containers)
	} else {
		slog.Debug("Docker istatistikleri alınırken hata oluştu", "err", err)
	}
	// Ek dosya sistemlerini ekleyin
	systemData.Stats.ExtraFs = make(map[string]*system.FsStats)
	for name, stats := range a.fsStats {
		if !stats.Root && stats.DiskTotal > 0 {
			systemData.Stats.ExtraFs[name] = stats
		}
	}
	slog.Debug("Ek dosya sistemleri", "data", systemData.Stats.ExtraFs)
	return systemData
}
