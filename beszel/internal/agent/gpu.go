package agent

import (
	"beszel/internal/entities/system"
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

// GPUManager, Nvidia veya AMD GPU'lar için veri toplama işlemlerini yönetir
type GPUManager struct {
	nvidiaSmi  bool
	rocmSmi    bool
	GpuDataMap map[string]*system.GPUData
	mutex      sync.Mutex
}

// RocmSmiJson, rocm-smi çıktısının JSON yapısını temsil eder
type RocmSmiJson struct {
	ID          string `json:"GUID"`
	Name        string `json:"Card series"`
	Temperature string `json:"Temperature (Sensor edge) (C)"`
	MemoryUsed  string `json:"VRAM Total Used Memory (B)"`
	MemoryTotal string `json:"VRAM Total Memory (B)"`
	Usage       string `json:"GPU use (%)"`
	Power       string `json:"Current Socket Graphics Package Power (W)"`
}

// gpuCollector, belirli bir GPU yönetim aracının (nvidia-smi veya rocm-smi) toplayıcısını tanımlar
type gpuCollector struct {
	name  string
	cmd   *exec.Cmd
	parse func([]byte) bool // geçerli veri bulunduğunda true döner
}

var errNoValidData = fmt.Errorf("geçerli GPU verisi bulunamadı") // Veri eksikliği hatası

// Belirtilen GPU yönetim aracı için veri toplama işlemini başlatır ve yönetir
func (c *gpuCollector) start() {
	for {
		err := c.collect()
		if err != nil {
			if err == errNoValidData {
				slog.Warn(c.name + " geçerli GPU verisi bulamadı, durduruluyor")
				break
			}
			slog.Warn(c.name+" başarısız oldu, yeniden başlatılıyor", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

// collect, komutu çalıştırır ve çıktıyı atanmış ayrıştırıcı fonksiyon ile ayrıştırır
func (c *gpuCollector) collect() error {
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := c.cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 8*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)

	hasValidData := false
	for scanner.Scan() {
		if c.parse(scanner.Bytes()) {
			hasValidData = true
		}
	}

	if !hasValidData {
		return errNoValidData
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("tarayıcı hatası: %w", err)
	}
	return c.cmd.Wait()
}

// parseNvidiaData, nvidia-smi çıktısını ayrıştırır ve GPUData haritasını günceller
func (gm *GPUManager) parseNvidiaData(output []byte) bool {
	fields := strings.Split(string(output), ", ")
	if len(fields) < 7 {
		return false
	}
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			fields := strings.Split(line, ", ")
			if len(fields) >= 7 {
				id := fields[0]
				temp, _ := strconv.ParseFloat(fields[2], 64)
				memoryUsage, _ := strconv.ParseFloat(fields[3], 64)
				totalMemory, _ := strconv.ParseFloat(fields[4], 64)
				usage, _ := strconv.ParseFloat(fields[5], 64)
				power, _ := strconv.ParseFloat(fields[6], 64)
				// GPU'yu ekle, eğer yoksa
				if _, ok := gm.GpuDataMap[id]; !ok {
					name := strings.TrimPrefix(fields[1], "NVIDIA ")
					gm.GpuDataMap[id] = &system.GPUData{Name: strings.TrimSuffix(name, " Laptop GPU")}
				}
				// GPU verisini güncelle
				gpu := gm.GpuDataMap[id]
				gpu.Temperature = temp
				gpu.MemoryUsed = memoryUsage / 1.024
				gpu.MemoryTotal = totalMemory / 1.024
				gpu.Usage += usage
				gpu.Power += power
				gpu.Count++
			}
		}
	}
	return true
}

// parseAmdData, rocm-smi çıktısını ayrıştırır ve GPUData haritasını günceller
func (gm *GPUManager) parseAmdData(output []byte) bool {
	var rocmSmiInfo map[string]RocmSmiJson
	if err := json.Unmarshal(output, &rocmSmiInfo); err != nil || len(rocmSmiInfo) == 0 {
		return false
	}
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	for _, v := range rocmSmiInfo {
		temp, _ := strconv.ParseFloat(v.Temperature, 64)
		memoryUsage, _ := strconv.ParseFloat(v.MemoryUsed, 64)
		totalMemory, _ := strconv.ParseFloat(v.MemoryTotal, 64)
		usage, _ := strconv.ParseFloat(v.Usage, 64)
		power, _ := strconv.ParseFloat(v.Power, 64)
		memoryUsage = bytesToMegabytes(memoryUsage)
		totalMemory = bytesToMegabytes(totalMemory)

		if _, ok := gm.GpuDataMap[v.ID]; !ok {
			gm.GpuDataMap[v.ID] = &system.GPUData{Name: v.Name}
		}
		gpu := gm.GpuDataMap[v.ID]
		gpu.Temperature = temp
		gpu.MemoryUsed = memoryUsage
		gpu.MemoryTotal = totalMemory
		gpu.Usage += usage
		gpu.Power += power
		gpu.Count++
	}
	return true
}

// Mevcut GPU kullanım verilerini toplar ve sıfırlar
func (gm *GPUManager) GetCurrentData() map[string]system.GPUData {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	// Aynı ada sahip GPU'ları kontrol et
	nameCounts := make(map[string]int)
	for _, gpu := range gm.GpuDataMap {
		nameCounts[gpu.Name]++
	}

	// Veriyi kopyala / sıfırla
	gpuData := make(map[string]system.GPUData, len(gm.GpuDataMap))
	for id, gpu := range gm.GpuDataMap {
		// Veriyi topla
		gpu.Temperature = twoDecimals(gpu.Temperature)
		gpu.MemoryUsed = twoDecimals(gpu.MemoryUsed)
		gpu.MemoryTotal = twoDecimals(gpu.MemoryTotal)
		gpu.Usage = twoDecimals(gpu.Usage / gpu.Count)
		gpu.Power = twoDecimals(gpu.Power / gpu.Count)
		// Sayacı sıfırla
		gpu.Count = 1
		// Başka bir şeyi üzerine yazmamak için referansı kaldır
		gpuCopy := *gpu
		// Aynı ada sahip birden fazla GPU varsa, isme id ekle
		if nameCounts[gpu.Name] > 1 {
			gpuCopy.Name = fmt.Sprintf("%s %s", gpu.Name, id)
		}
		gpuData[id] = gpuCopy
	}
	return gpuData
}

// detectGPUs, GPU markasını (nvidia veya amd) döner veya bulunamazsa hata döner
// todo: Gerçekten bir GPU olup olmadığını kontrol et, sadece komutun var olup olmadığını değil
func (gm *GPUManager) detectGPUs() error {
	if err := exec.Command("nvidia-smi").Run(); err == nil {
		gm.nvidiaSmi = true
	}
	if err := exec.Command("rocm-smi").Run(); err == nil {
		gm.rocmSmi = true
	}
	if gm.nvidiaSmi || gm.rocmSmi {
		return nil
	}
	return fmt.Errorf("GPU bulunamadı - nvidia-smi veya rocm-smi yükleyin")
}

// startCollector, komuta bağlı olarak uygun GPU veri toplayıcısını başlatır
func (gm *GPUManager) startCollector(command string) {
	switch command {
	case "nvidia-smi":
		nvidia := gpuCollector{
			name: "nvidia-smi",
			cmd: exec.Command("nvidia-smi", "-l", "4",
				"--query-gpu=index,name,temperature.gpu,memory.used,memory.total,utilization.gpu,power.draw",
				"--format=csv,noheader,nounits"),
			parse: gm.parseNvidiaData,
		}
		go nvidia.start()
	case "rocm-smi":
		amdCollector := gpuCollector{
			name: "rocm-smi",
			cmd: exec.Command("/bin/sh", "-c",
				"while true; do rocm-smi --showid --showtemp --showuse --showpower --showproductname --showmeminfo vram --json; sleep 4.3; done"),
			parse: gm.parseAmdData,
		}
		go amdCollector.start()
	}
}

// NewGPUManager, yeni bir GPUManager oluşturur ve başlatır
func NewGPUManager() (*GPUManager, error) {
	var gm GPUManager
	if err := gm.detectGPUs(); err != nil {
		return nil, err
	}
	gm.GpuDataMap = make(map[string]*system.GPUData, 1)

	if gm.nvidiaSmi {
		gm.startCollector("nvidia-smi")
	}
	if gm.rocmSmi {
		gm.startCollector("rocm-smi")
	}

	return &gm, nil
}
