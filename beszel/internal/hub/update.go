package hub

import (
	"beszel"
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

// Update beszel'i en son sürüme günceller
func Update(_ *cobra.Command, _ []string) {
	var latest *selfupdate.Release
	var found bool
	var err error
	currentVersion := semver.MustParse(beszel.Version)
	fmt.Println("beszel", currentVersion)
	fmt.Println("Güncellemeler kontrol ediliyor...")
	updater, _ := selfupdate.NewUpdater(selfupdate.Config{
		Filters: []string{"beszel_"},
	})
	latest, found, err = updater.DetectLatest("henrygd/beszel")

	if err != nil {
		fmt.Println("Güncellemeleri kontrol ederken hata oluştu:", err)
		os.Exit(1)
	}

	if !found {
		fmt.Println("Güncelleme bulunamadı")
		os.Exit(0)
	}

	fmt.Println("En son sürüm:", latest.Version)

	if latest.Version.LTE(currentVersion) {
		fmt.Println("Güncelsiniz")
		return
	}

	var binaryPath string
	fmt.Printf("%s sürümünden %s sürümüne güncelleniyor...\n", currentVersion, latest.Version)
	binaryPath, err = os.Executable()
	if err != nil {
		fmt.Println("Binary yolunu alırken hata oluştu:", err)
		os.Exit(1)
	}
	err = selfupdate.UpdateTo(latest.AssetURL, binaryPath)
	if err != nil {
		fmt.Println("Lütfen sudo ile tekrar deneyin. Hata:", err)
		os.Exit(1)
	}
	fmt.Printf("Başarıyla %s sürümüne güncellendi\n\n%s\n", latest.Version, strings.TrimSpace(latest.ReleaseNotes))
}
