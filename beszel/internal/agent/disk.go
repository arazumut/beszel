package agent

import (
	"beszel/internal/entities/system"
	"log/slog"
	"time"

	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

// Disk kullanımı ve I/O için izlenecek dosya sistemlerini ayarlar.
func (a *Agent) initializeDiskInfo() {
	filesystem := os.Getenv("FILESYSTEM")
	efPath := "/extra-filesystems"
	hasRoot := false

	partitions, err := disk.Partitions(false)
	if err != nil {
		slog.Error("Disk bölümlerini alırken hata oluştu", "err", err)
	}
	slog.Debug("Disk", "partitions", partitions)

	diskIoCounters, err := disk.IOCounters()
	if err != nil {
		slog.Error("Disk istatistiklerini alırken hata oluştu", "err", err)
	}
	slog.Debug("Disk I/O", "diskstats", diskIoCounters)

	// Dosya sistemini fsStats'a eklemek için yardımcı fonksiyon
	addFsStat := func(device, mountpoint string, root bool) {
		key := filepath.Base(device)
		var ioMatch bool
		if _, exists := a.fsStats[key]; !exists {
			if root {
				slog.Info("Kök cihaz tespit edildi", "name", key)
				// Kök cihazın /proc/diskstats içinde olup olmadığını kontrol et, değilse yedek kullan
				if _, ioMatch = diskIoCounters[key]; !ioMatch {
					key, ioMatch = findIoDevice(filesystem, diskIoCounters, a.fsStats)
					if !ioMatch {
						slog.Info("I/O yedeği kullanılıyor", "device", device, "mountpoint", mountpoint, "fallback", key)
					}
				}
			} else {
				// Kök olmayan cihazın disk istatistiklerinde olup olmadığını kontrol et, değilse klasör adını kullan
				if _, ioMatch = diskIoCounters[key]; !ioMatch {
					efBase := filepath.Base(mountpoint)
					if _, ioMatch = diskIoCounters[efBase]; ioMatch {
						key = efBase
					}
				}
			}
			a.fsStats[key] = &system.FsStats{Root: root, Mountpoint: mountpoint}
		}
	}

	// FILESYSTEM ortam değişkenini kullanarak kök dosya sistemini bul
	if filesystem != "" {
		for _, p := range partitions {
			if strings.HasSuffix(p.Device, filesystem) || p.Mountpoint == filesystem {
				addFsStat(p.Device, p.Mountpoint, true)
				hasRoot = true
				break
			}
		}
		if !hasRoot {
			slog.Warn("Bölüm detayları bulunamadı", "filesystem", filesystem)
		}
	}

	// EXTRA_FILESYSTEMS ortam değişkeni değerlerini fsStats'a ekle
	if extraFilesystems, exists := os.LookupEnv("EXTRA_FILESYSTEMS"); exists {
		for _, fs := range strings.Split(extraFilesystems, ",") {
			found := false
			for _, p := range partitions {
				if strings.HasSuffix(p.Device, fs) || p.Mountpoint == fs {
					addFsStat(p.Device, p.Mountpoint, false)
					found = true
					break
				}
			}
			// bölümlerde değilse, disk kullanımını alıp alamayacağımızı test et
			if !found {
				if _, err := disk.Usage(fs); err == nil {
					addFsStat(filepath.Base(fs), fs, false)
				} else {
					slog.Error("Geçersiz dosya sistemi", "name", fs, "err", err)
				}
			}
		}
	}

	// Çeşitli montaj noktaları için bölümleri işle
	for _, p := range partitions {
		// İkili kök yedeği veya docker kök yedeği
		if !hasRoot && (p.Mountpoint == "/" || (p.Mountpoint == "/etc/hosts" && strings.HasPrefix(p.Device, "/dev"))) {
			fs, match := findIoDevice(filepath.Base(p.Device), diskIoCounters, a.fsStats)
			if match {
				addFsStat(fs, p.Mountpoint, true)
				hasRoot = true
			}
		}

		// Cihazın /extra-filesystems içinde olup olmadığını kontrol et
		if strings.HasPrefix(p.Mountpoint, efPath) {
			addFsStat(p.Device, p.Mountpoint, false)
		}
	}

	// /extra-filesystems içindeki tüm klasörleri kontrol et ve henüz eklenmemişse ekle
	if folders, err := os.ReadDir(efPath); err == nil {
		existingMountpoints := make(map[string]bool)
		for _, stats := range a.fsStats {
			existingMountpoints[stats.Mountpoint] = true
		}
		for _, folder := range folders {
			if folder.IsDir() {
				mountpoint := filepath.Join(efPath, folder.Name())
				slog.Debug("/extra-filesystems", "mountpoint", mountpoint)
				if !existingMountpoints[mountpoint] {
					addFsStat(folder.Name(), mountpoint, false)
				}
			}
		}
	}

	// Kök dosya sistemi ayarlanmadıysa, yedek kullan
	if !hasRoot {
		rootDevice, _ := findIoDevice(filepath.Base(filesystem), diskIoCounters, a.fsStats)
		slog.Info("Kök disk", "mountpoint", "/", "io", rootDevice)
		a.fsStats[rootDevice] = &system.FsStats{Root: true, Mountpoint: "/"}
	}

	a.initializeDiskIoStats(diskIoCounters)
}

// /proc/diskstats içinden eşleşen cihazı döndürür,
// veya eşleşme bulunamazsa en çok okuma yapan cihazı döndürür.
// bool, bir eşleşme bulunursa true döner.
func findIoDevice(filesystem string, diskIoCounters map[string]disk.IOCountersStat, fsStats map[string]*system.FsStats) (string, bool) {
	var maxReadBytes uint64
	maxReadDevice := "/"
	for _, d := range diskIoCounters {
		if d.Name == filesystem || (d.Label != "" && d.Label == filesystem) {
			return d.Name, true
		}
		if d.ReadBytes > maxReadBytes {
			// cihaz zaten fsStats içinde varsa kullanma
			if _, exists := fsStats[d.Name]; !exists {
				maxReadBytes = d.ReadBytes
				maxReadDevice = d.Name
			}
		}
	}
	return maxReadDevice, false
}

// Disk I/O istatistikleri için başlangıç değerlerini ayarlar.
func (a *Agent) initializeDiskIoStats(diskIoCounters map[string]disk.IOCountersStat) {
	for device, stats := range a.fsStats {
		// diskIoCounters içinde değilse atla
		d, exists := diskIoCounters[device]
		if !exists {
			slog.Warn("Cihaz disk istatistiklerinde bulunamadı", "name", device)
			continue
		}
		// başlangıç değerlerini doldur
		stats.Time = time.Now()
		stats.TotalRead = d.ReadBytes
		stats.TotalWrite = d.WriteBytes
		// geçerli io cihaz adları listesine ekle
		a.fsNames = append(a.fsNames, device)
	}
}
