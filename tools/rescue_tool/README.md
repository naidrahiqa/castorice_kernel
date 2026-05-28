# 🚨 Epitaph Rescue Tool

[![Go Version](https://img.shields.io/github/go-mod/go-version/naidrahiqa/epitaph_resque?color=6366f1&style=for-the-badge)](https://go.dev)
[![Fyne Version](https://img.shields.io/badge/UI-Fyne%20v2.7.4-06b6d4?style=for-the-badge)](https://fyne.io)
[![Platform Support](https://img.shields.io/badge/Platform-Windows-blue?style=for-the-badge&logo=windows&logoColor=white)](https://github.com/naidrahiqa/epitaph_resque)
[![Build Status](https://img.shields.io/github/actions/workflow/status/naidrahiqa/epitaph_resque/build.yml?branch=main&style=for-the-badge&logo=github)](https://github.com/naidrahiqa/epitaph_resque/actions)
[![Latest Release](https://img.shields.io/github/v/release/naidrahiqa/epitaph_resque?color=22c55e&style=for-the-badge)](https://github.com/naidrahiqa/epitaph_resque/releases)

**Epitaph Rescue Tool** adalah utility pemulihan dan debug kernel berbasis GUI yang dirancang khusus untuk pengguna **Redmi 12 (fire)** yang mengalami bootloop setelah flashing custom kernel (khususnya **Epitaph Kernel**). Dibuat menggunakan **Go 1.26** dan framework **Fyne v2** untuk menghasilkan Windows executable portabel berukuran kecil dan berkinerja tinggi.

---

## 🔍 Mengapa Tool Ini Dibuat? (Problem Solved)
Flashing kernel baru rentan menimbulkan masalah bootloop bagi pengguna awam. Ketika HP mereka bootloop, mereka seringkali panik dan kesulitan:
1. Melakukan deteksi status device (Fastboot vs Android).
2. Mem-flash stock `boot.img` cadangan sebelum melakukan ekstraksi log.
3. Menarik (pull) log crash kernel (`console-ramoops` / `last_kmsg`) dari partisi PStore.
4. Menganalisis penyebab crash secara manual karena tidak paham kode error kernel.
5. Menyelesaikan masalah WiFi mati setelah flash rescue kernel via Fastboot (karena modul vendor mismatch).

**Epitaph Rescue Tool** menyelesaikan semua masalah di atas melalui wizard interaktif **step-by-step** yang ramah pengguna.

---

## 🛠️ Fitur Utama

### 1. 📱 Home & Device Discovery
* Auto-polling status koneksi setiap **2 detik**.
* Deteksi mode secara dinamis: **Android (ADB)**, **Fastboot Mode**, atau **Disconnected**.
* Menampilkan info spesifikasi device secara mendalam: LCM Panel variant, Kernel Version, active KernelSU (KSU) version, active SUSFS version, Android SDK version, dan ROM build ID.
* **Platform-Tools Auto-Downloader**: Mendownload SDK tools resmi secara aman langsung dari server Google dan mengekstraknya ke `%APPDATA%\EpitaphRescue` jika sistem belum memilikinya.

### 2. 🚨 Guided Rescue Wizard
* Panduan pemulihan interaktif terstruktur:
  1. **Detect**: Otomatis mendeteksi situasi koneksi hardware.
  2. **Flash**: File picker interaktif untuk memilih stock `boot.img` dengan verifikasi otomatis **magic bytes** (`ANDROID!`) demi keselamatan flashing.
  3. **Reboot**: Memboot ulang sistem secara otomatis.
  4. **Pull Log**: Menarik log debug dari filesystem RAMoops.
  5. **Analyze**: Menyajikan diagnosis otomatis masalah terpopuler.

### 3. 📋 Log Parser & Search Terminal
* **Syntax Highlighting**: Mewarnai baris log crash berdasarkan tingkat keparahan (Merah untuk `Kernel Panic` / OOM, Kuning untuk `avc denied` SELinux, Cyan untuk `KSU`, Hijau untuk `SUSFS`).
* **Live Query Filtering**: Memfilter ratusan baris log secara instan menggunakan kata kunci.
* **Developer Quick-Actions**: Copy log ke clipboard sistem atau buka secara instan di Windows `Notepad` untuk keperluan share log ke grup Telegram.

### 4. 📶 WiFi Fixer Guide
* Diagnosa wlan network interface (`wlan0`), driver mod (`lsmod`), dan mount point `vendor_dlkm`.
* Panduan pemecahan masalah interaktif untuk mem-flash stock boot dan meng-update modul kernel melalui *KernelFlasher* (menghindari WiFi mati karena mismatch flash via Fastboot).

### 5. ✅ Pre-Flash Preflight Validator
* Menguji kepatuhan device sebelum user flash kernel untuk pertama kali: codename verification (`fire`), Android version, vbmeta/verified boot state (`orange`), LCD panel warning (`LC0C/LC0D`), dan ROM build.

---

## 🎨 Design System & Palette

Antarmuka dirancang dengan estetika bertema **Premium Dark / Indigo** yang serasi dan konsisten untuk kenyamanan mata pengguna:

*   **Deep Backdrop Background:** `#0d1117`
*   **Card Surface:** `#161b22`
*   **Primary Accent:** `#6366f1` (Indigo)
*   **Focus / Selection Glow:** `#06b6d4` (Cyan)
*   **Success Indicator:** `#22c55e` (Green)
*   **Warning Alert:** `#f59e0b` (Amber/Orange)
*   **Critical Danger:** `#ef4444` (Red)

---

## 🚀 Cara Kompilasi (Build Guide)

### Persyaratan:
1. **Go 1.22+** terinstall.
2. **C compiler (gcc/MinGW-w64)** terinstall di PATH Windows Anda (dibutuhkan oleh OpenGL/GLFW pada Fyne framework).

### Langkah Build Manual:
Jalankan di command line (PowerShell/CMD):
```powershell
# Set CGO compiler flag aktif
$env:CGO_ENABLED="1"

# Kompilasi executable portabel tanpa command terminal di latar belakang (-H windowsgui)
go build -ldflags="-H windowsgui -s -w" -o EpitaphRescue.exe .
```

---

## 👤 Mengapa Commit / Author Masih Menggunakan "AI Assistant"?

Jika Anda melihat kontribusi atau riwayat git log repository ini tercatat sebagai **AI Assistant** (atau akun default gemini), hal ini terjadi karena **Git lokal Anda belum dikonfigurasi dengan username dan email akun GitHub Anda**.

Untuk membetulkannya agar semua commit berikutnya tercatat atas nama Anda (**naidrahiqa**), silakan jalankan command berikut di terminal proyek:

```powershell
# Konfigurasi identitas Git global Anda
git config --global user.name "username-kamu"
git config --global user.email "email@kamu.com" # Ganti dengan email GitHub kamu

# Jika ingin mengubah nama pembuat commit TERAKHIR yang baru saja dibuat
git commit --amend --reset-author --no-edit
```
Setelah menjalankan perintah di atas, lakukan push ulang ke repositori GitHub untuk memperbarui identitas kontributor Anda secara profesional!

---

## 🤝 Kontribusi & Lisensi
Dikembangkan oleh **@naidrahiqa** untuk komunitas modifikasi Redmi 12 (fire).  
*Silakan ajukan Issue atau Pull Request jika Anda menemukan bug log signature baru.*
