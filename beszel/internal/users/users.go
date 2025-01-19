// Package users handles user-related custom functionality.
// Paket users, kullanıcıyla ilgili özel işlevselliği ele alır.
package users

import (
	"beszel/migrations"
	"log"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// UserManager, kullanıcı yönetimi için bir yapı tanımlar.
type UserManager struct {
	app *pocketbase.PocketBase
}

// UserSettings, kullanıcı ayarlarını temsil eder.
type UserSettings struct {
	ChartTime            string   `json:"chartTime"`
	NotificationEmails   []string `json:"emails"`
	NotificationWebhooks []string `json:"webhooks"`
}

// NewUserManager, yeni bir UserManager örneği oluşturur.
func NewUserManager(app *pocketbase.PocketBase) *UserManager {
	return &UserManager{
		app: app,
	}
}

// Kullanıcı rolünü başlatır, eğer ayarlanmamışsa varsayılan olarak "user" yapar.
func (um *UserManager) InitializeUserRole(e *core.RecordEvent) error {
	if e.Record.GetString("role") == "" {
		e.Record.Set("role", "user")
	}
	return e.Next()
}

// Kullanıcı ayarlarını varsayılanlarla başlatır, eğer ayarlanmamışsa.
func (um *UserManager) InitializeUserSettings(e *core.RecordEvent) error {
	record := e.Record
	// Ayarları varsayılanlarla başlat
	settings := UserSettings{
		ChartTime:            "1h",
		NotificationEmails:   []string{},
		NotificationWebhooks: []string{},
	}
	record.UnmarshalJSONField("settings", &settings)
	if len(settings.NotificationEmails) == 0 {
		// Kullanıcı e-postasını kimlik doğrulama kaydından al
		if errs := um.app.ExpandRecord(record, []string{"user"}, nil); len(errs) == 0 {
			if user := record.ExpandedOne("user"); user != nil {
				settings.NotificationEmails = []string{user.GetString("email")}
			} else {
				log.Println("Kimlik doğrulama kaydından kullanıcı e-postası alınamadı")
			}
		} else {
			log.Println("Kullanıcı ilişkisi genişletilemedi", "hatalar", errs)
		}
	}
	record.Set("settings", settings)
	return e.Next()
}

// İlk kullanıcıyı oluşturmak için özel bir API uç noktası.
// PocketBase < 0.23.0'deki önceki varsayılan davranışı taklit eder ve kullanıcının Beszel UI aracılığıyla oluşturulmasına izin verir.
func (um *UserManager) CreateFirstUser(e *core.RequestEvent) error {
	// Kullanıcı olmadığını kontrol et
	totalUsers, err := um.app.CountRecords("users")
	if err != nil || totalUsers > 0 {
		return e.JSON(http.StatusForbidden, map[string]string{"err": "Yasak"})
	}
	// Sadece bir süper kullanıcı olduğunu ve e-postanın initial-settings.go'da ayarladığımız süper kullanıcının e-postasıyla eşleştiğini kontrol et
	adminUsers, err := um.app.FindAllRecords(core.CollectionNameSuperusers)
	if err != nil || len(adminUsers) != 1 || adminUsers[0].GetString("email") != migrations.TempAdminEmail {
		return e.JSON(http.StatusForbidden, map[string]string{"err": "Yasak"})
	}
	// İstek gövdesinde sağlanan e-posta ve şifreyi kullanarak ilk kullanıcıyı oluştur
	data := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := e.BindBody(&data); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"err": err.Error()})
	}
	if data.Email == "" || data.Password == "" {
		return e.JSON(http.StatusBadRequest, map[string]string{"err": "Kötü istek"})
	}

	collection, _ := um.app.FindCollectionByNameOrId("users")
	user := core.NewRecord(collection)
	user.SetEmail(data.Email)
	user.SetPassword(data.Password)
	user.Set("role", "admin")
	user.Set("verified", true)
	if username := strings.Split(data.Email, "@")[0]; len(username) > 2 {
		user.Set("username", username)
	}
	if err := um.app.Save(user); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
	}
	// İlk kullanıcının e-postasını kullanarak süper kullanıcı oluştur
	collection, _ = um.app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	adminUser := core.NewRecord(collection)
	adminUser.SetEmail(data.Email)
	adminUser.SetPassword(data.Password)
	if err := um.app.Save(adminUser); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
	}
	// İlk süper kullanıcıyı sil
	if err := um.app.Delete(adminUsers[0]); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
	}
	return e.JSON(http.StatusOK, map[string]string{"msg": "Kullanıcı oluşturuldu"})
}
