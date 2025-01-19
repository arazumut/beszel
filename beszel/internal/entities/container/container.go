package container

import "time"

// /containers/json'dan Docker konteyner bilgisi
type ApiInfo struct {
	Id      string   // Konteyner ID'si
	IdShort string   // Kısa Konteyner ID'si
	Names   []string // Konteyner isimleri
	Status  string   // Konteyner durumu
}

// /containers/{id}/stats'dan Docker konteyner kaynakları
type ApiStats struct {
	Networks    map[string]NetworkStats // Ağ istatistikleri
	CPUStats    CPUStats                // CPU istatistikleri
	MemoryStats MemoryStats             // Bellek istatistikleri
}

type CPUStats struct {
	CPUUsage    CPUUsage // CPU Kullanımı
	SystemUsage uint64   // Sistem Kullanımı (sadece Linux)
}

type CPUUsage struct {
	TotalUsage uint64 // Toplam CPU kullanımı
}

type MemoryStats struct {
	Usage uint64           // Mevcut bellek kullanımı
	Stats MemoryStatsStats // Bellek istatistikleri
}

type MemoryStatsStats struct {
	Cache        uint64 // Önbellek
	InactiveFile uint64 // Pasif dosya
}

type NetworkStats struct {
	RxBytes uint64 // Alınan baytlar
	TxBytes uint64 // Gönderilen baytlar
}

type prevNetStats struct {
	Sent uint64    // Gönderilen baytlar
	Recv uint64    // Alınan baytlar
	Time time.Time // Zaman
}

// Docker konteyner istatistikleri
type Stats struct {
	Name        string       `json:"n"`  // İsim
	Cpu         float64      `json:"c"`  // CPU kullanımı
	Mem         float64      `json:"m"`  // Bellek kullanımı
	NetworkSent float64      `json:"ns"` // Gönderilen ağ verisi
	NetworkRecv float64      `json:"nr"` // Alınan ağ verisi
	PrevCpu     [2]uint64    `json:"-"`  // Önceki CPU kullanımı
	PrevNet     prevNetStats `json:"-"`  // Önceki ağ istatistikleri
}
