package agent

import (
	"log/slog"
	"os"
	"strings"
	"time"

	psutilNet "github.com/shirou/gopsutil/v4/net"
)

func (a *Agent) initializeNetIoStats() {
	// Geçerli ağ arayüzlerini sıfırla
	a.netInterfaces = make(map[string]struct{}, 0)

	// NICS ortam değişkeni ile iletilen ağ arayüzü isimlerinin haritası
	var nicsMap map[string]struct{}
	nics, nicsEnvExists := os.LookupEnv("NICS")
	if nicsEnvExists {
		nicsMap = make(map[string]struct{}, 0)
		for _, nic := range strings.Split(nics, ",") {
			nicsMap[nic] = struct{}{}
		}
	}

	// Ağ I/O istatistiklerini sıfırla
	a.netIoStats.BytesSent = 0
	a.netIoStats.BytesRecv = 0

	// İlk ağ I/O istatistiklerini al
	if netIO, err := psutilNet.IOCounters(true); err == nil {
		a.netIoStats.Time = time.Now()
		for _, v := range netIO {
			switch {
			// Eğer nics varsa ve arayüz listede yoksa atla
			case nicsEnvExists:
				if _, nameInNics := nicsMap[v.Name]; !nameInNics {
					continue
				}
			// Aksi takdirde arayüz ismini skipNetworkInterface fonksiyonundan geçir
			default:
				if a.skipNetworkInterface(v) {
					continue
				}
			}
			slog.Info("Ağ arayüzü tespit edildi", "isim", v.Name, "gönderilen", v.BytesSent, "alınan", v.BytesRecv)
			a.netIoStats.BytesSent += v.BytesSent
			a.netIoStats.BytesRecv += v.BytesRecv
			// Geçerli bir ağ arayüzü olarak sakla
			a.netInterfaces[v.Name] = struct{}{}
		}
	}
}

func (a *Agent) skipNetworkInterface(v psutilNet.IOCountersStat) bool {
	switch {
	case strings.HasPrefix(v.Name, "lo"),
		strings.HasPrefix(v.Name, "docker"),
		strings.HasPrefix(v.Name, "br-"),
		strings.HasPrefix(v.Name, "veth"),
		v.BytesRecv == 0,
		v.BytesSent == 0:
		return true
	default:
		return false
	}
}
