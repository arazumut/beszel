# Beszel

Beszel is a lightweight server monitoring platform that includes Docker statistics, historical data, and alert functions.

It has a friendly web interface, simple configuration, and is ready to use out of the box. It supports automatic backup, multi-user, OAuth authentication, and API access.

[![agent Docker Image Size](https://img.shields.io/docker/image-size/henrygd/beszel-agent/0.4.0?logo=docker&label=agent%20image%20size)](https://hub.docker.com/r/henrygd/beszel-agent)
[![hub Docker Image Size](https://img.shields.io/docker/image-size/henrygd/beszel/0.4.0?logo=docker&label=hub%20image%20size)](https://hub.docker.com/r/henrygd/beszel)
[![MIT license](https://img.shields.io/github/license/henrygd/beszel?color=%239944ee)](https://github.com/henrygd/beszel/blob/main/LICENSE)
[![Crowdin](https://badges.crowdin.net/beszel/localized.svg)](https://crowdin.com/project/beszel)

![Screenshot of beszel dashboard and system page](https://henrygd-assets.b-cdn.net/beszel/screenshot-new.png)

## Features

- **Lightweight**: Smaller and less resource-intensive than leading solutions.
- **Simple**: Easy setup, no need for public internet exposure.
- **Docker stats**: Tracks CPU, memory, and network usage history for each container.
- **Alerts**: Configurable alerts for CPU, memory, disk, bandwidth, temperature, and status.
- **Multi-user**: Users manage their own systems. Admins can share systems across users.
- **OAuth / OIDC**: Supports many OAuth2 providers. Password auth can be disabled.
- **Automatic backups**: Save and restore data from disk or S3-compatible storage.
- **REST API**: Use or update your data in your own scripts and applications.

## Architecture

Beszel consists of two main components: the **hub** and the **agent**.

- **Hub**: A web application built on [PocketBase](https://pocketbase.io/) that provides a dashboard for viewing and managing connected systems.
- **Agent**: Runs on each system you want to monitor, creating a minimal SSH server to communicate system metrics to the hub.

## Getting started

The [quick start guide](https://beszel.dev/guide/getting-started) and other documentation is available on our website, [beszel.dev](https://beszel.dev). You'll be up and running in a few minutes.

## Screenshots

![Dashboard](https://beszel.dev/image/dashboard.png)
![System page](https://beszel.dev/image/system-full.png)
![Notification Settings](https://beszel.dev/image/settings-notifications.png)

## Supported metrics

- **CPU usage** - Host system and Docker / Podman containers.
- **Memory usage** - Host system and containers. Includes swap and ZFS ARC.
- **Disk usage** - Host system. Supports multiple partitions and devices.
- **Disk I/O** - Host system. Supports multiple partitions and devices.
- **Network usage** - Host system and containers.
- **Temperature** - Host system sensors.
- **GPU usage / temperature / power draw** - Nvidia and AMD only. Must use binary agent.

## License

Beszel is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.







----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------








Beszel
Beszel, Docker istatistikleri, geçmiş veriler ve uyarı işlevleri içeren hafif bir sunucu izleme platformudur.

Kullanıcı dostu bir web arayüzüne, basit bir yapılandırmaya sahiptir ve kutudan çıkar çıkmaz kullanıma hazırdır. Otomatik yedekleme, çoklu kullanıcı desteği, OAuth kimlik doğrulama ve API erişimi gibi özellikleri destekler.








Özellikler
Hafif: Lider çözümlere göre daha küçük ve daha az kaynak tüketir.
Basit: Kolay kurulum, genel internet erişimi gerekmez.
Docker istatistikleri: Her konteyner için CPU, bellek ve ağ kullanım geçmişini izler.
Uyarılar: CPU, bellek, disk, bant genişliği, sıcaklık ve durum için yapılandırılabilir uyarılar.
Çoklu kullanıcı: Kullanıcılar kendi sistemlerini yönetebilir. Yöneticiler sistemleri kullanıcılar arasında paylaşabilir.
OAuth / OIDC: Birçok OAuth2 sağlayıcısını destekler. Şifre ile kimlik doğrulama devre dışı bırakılabilir.
Otomatik yedekleme: Verileri disk veya S3 uyumlu depolama alanına kaydedip geri yükleyin.
REST API: Verilerinizi kendi betikleriniz ve uygulamalarınızda kullanın veya güncelleyin.
Mimari
Beszel, iki ana bileşenden oluşur: hub ve agent.

Hub: Bağlı sistemleri görüntülemek ve yönetmek için bir pano sağlayan PocketBase tabanlı bir web uygulamasıdır.
Agent: İzlemek istediğiniz her sistemde çalışır ve sistem metriklerini hub’a iletmek için minimal bir SSH sunucusu oluşturur.
Başlarken
Hızlı başlangıç kılavuzu ve diğer belgeler web sitemizde mevcuttur: beszel.dev. Birkaç dakika içinde kurulumu tamamlayabilirsiniz.

Ekran Görüntüleri




Desteklenen Metrikler
CPU kullanımı - Ana sistem ve Docker / Podman konteynerleri.
Bellek kullanımı - Ana sistem ve konteynerler. Swap ve ZFS ARC dahil.
Disk kullanımı - Ana sistem. Birden fazla bölüm ve cihaz destekler.
Disk G/Ç - Ana sistem. Birden fazla bölüm ve cihaz destekler.
Ağ kullanımı - Ana sistem ve konteynerler.
Sıcaklık - Ana sistem sensörleri.
GPU kullanımı / sıcaklık / güç tüketimi - Sadece Nvidia ve AMD. İkili agent kullanılması gerekir.
Lisans
Beszel, MIT Lisansı ile lisanslanmıştır. Daha fazla ayrıntı için LİSANS dosyasına bakın.
