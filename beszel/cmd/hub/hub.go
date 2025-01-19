package main

import (
	"beszel"
	"beszel/internal/hub"

	_ "beszel/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

func main() {
	// Yeni bir PocketBase uygulaması oluştur
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: beszel.AppName + "_data", // Varsayılan veri dizini
	})
	app.RootCmd.Version = beszel.Version // Uygulama versiyonunu ayarla
	app.RootCmd.Use = beszel.AppName     // Uygulama adını ayarla
	app.RootCmd.Short = ""               // Kısa açıklama (boş bırakılmış)

	// Güncelleme komutunu ekle
	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: beszel.AppName + " uygulamasını en son sürüme güncelle",
		Run:   hub.Update,
	})

	// Yeni bir Hub oluştur ve çalıştır
	hub.NewHub(app).Run()
}
