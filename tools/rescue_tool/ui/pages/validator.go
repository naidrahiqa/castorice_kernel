package pages

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/validator"
)

type FlasherPage struct {
	deviceMgr   *adb.DeviceManager
	window      fyne.Window
	val         *validator.Validator

	// Flasher UI
	selectedFilePath string
	lblSelectedFile  *widget.Label
	btnSelectFile    *widget.Button
	selectPartition  *widget.Select
	btnFlash         *widget.Button

	// Backup UI
	selectBackupPart *widget.Select
	btnBackup        *widget.Button

	// Validator UI
	btnCheck         *widget.Button
	lblSummary       *widget.Label
	resultBox        *fyne.Container

	mainCard         *widget.Card
	mainBox          *fyne.Container
}

func NewFlasherPage(dm *adb.DeviceManager, w fyne.Window) *FlasherPage {
	fp := &FlasherPage{
		deviceMgr: dm,
		window:    w,
		val:       validator.NewValidator(dm),
	}
	fp.buildUI()
	return fp
}

func (fp *FlasherPage) buildUI() {
	// 1. FLASHER UI
	fp.lblSelectedFile = widget.NewLabel("Belum ada file .img yang dipilih")
	fp.lblSelectedFile.Wrapping = fyne.TextWrapWord

	fp.btnSelectFile = widget.NewButtonWithIcon("Pilih File Partition Image (.img)", theme.FolderOpenIcon(), func() {
		fp.selectImageFile()
	})

	partitions := []string{"boot", "vendor_boot", "init_boot", "recovery", "dtbo", "vbmeta"}
	fp.selectPartition = widget.NewSelect(partitions, func(s string) {})
	fp.selectPartition.SetSelected("boot")

	fp.btnFlash = widget.NewButtonWithIcon("⚡ Flash Partition via Fastboot", theme.ConfirmIcon(), func() {
		fp.runFlash()
	})
	fp.btnFlash.Importance = widget.HighImportance

	flasherForm := widget.NewForm(
		widget.NewFormItem("1. Pilih File", container.NewVBox(fp.btnSelectFile, fp.lblSelectedFile)),
		widget.NewFormItem("2. Pilih Partisi Target", fp.selectPartition),
	)

	flasherCard := NewNeoCard("Universal Partition Flasher", "Flash .img ke partisi HP via Fastboot", container.NewVBox(
		flasherForm,
		NeoDivider(),
		fp.btnFlash,
	))

	// 2. BACKUP UI
	fp.selectBackupPart = widget.NewSelect(partitions, func(s string) {})
	fp.selectBackupPart.SetSelected("boot")

	fp.btnBackup = widget.NewButtonWithIcon("💾 Backup Partition via ADB (Root)", theme.DownloadIcon(), func() {
		fp.runBackup()
	})
	fp.btnBackup.Importance = widget.SuccessImportance

	backupForm := widget.NewForm(
		widget.NewFormItem("Partisi Backup", fp.selectBackupPart),
	)

	backupCard := NewNeoCard("Partition Backup & Dumper", "Dump partisi ke folder PC (butuh Root)", container.NewVBox(
		backupForm,
		NeoDivider(),
		fp.btnBackup,
	))

	// 3. VALIDATOR UI
	fp.btnCheck = widget.NewButtonWithIcon("Mulai Pre-Flash Check", theme.ConfirmIcon(), func() {
		fp.runChecks()
	})
	fp.btnCheck.Importance = widget.HighImportance

	fp.lblSummary = widget.NewLabel("Silakan hubungkan device dalam Android mode untuk memvalidasi spesifikasi hardware & verified boot.")
	fp.lblSummary.Wrapping = fyne.TextWrapWord

	fp.resultBox = container.NewVBox()

	validatorCard := NewNeoCard("Pre-Flash Device Validator", "Verifikasi spesifikasi sebelum flashing", container.NewVBox(
		fp.lblSummary,
		fp.btnCheck,
		NeoDivider(),
		fp.resultBox,
	))

	// Main Layout
	fp.mainBox = container.NewVBox(
		flasherCard,
		widget.NewSeparator(),
		backupCard,
		widget.NewSeparator(),
		validatorCard,
	)
}

func (fp *FlasherPage) Content() fyne.CanvasObject {
	return container.NewScroll(fp.mainBox)
}

func (fp *FlasherPage) selectImageFile() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		fp.selectedFilePath = reader.URI().Path()
		fp.lblSelectedFile.SetText(fp.selectedFilePath)
	}, fp.window)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".img"}))
	fd.Show()
}

func (fp *FlasherPage) runFlash() {
	if fp.selectedFilePath == "" {
		dialog.ShowError(fmt.Errorf("silakan pilih file .img terlebih dahulu."), fp.window)
		return
	}

	info := fp.deviceMgr.Detect()
	if info.Mode != adb.ModeFastboot {
		dialog.ShowError(fmt.Errorf("perangkat harus terhubung dalam mode Fastboot untuk mem-flash. Silakan reboot ke Fastboot di tab Home."), fp.window)
		return
	}

	part := fp.selectPartition.Selected
	confirmMsg := fmt.Sprintf("Apakah Anda yakin ingin mem-flash file:\n%s\n\nKe partisi target: %s?\nTindakan salah dapat mengakibatkan HP bootloop/hard-brick.", fp.selectedFilePath, part)

	dialog.ShowConfirm("Konfirmasi Flash", confirmMsg, func(ok bool) {
		if !ok {
			return
		}

		progress := dialog.NewCustomWithoutButtons("Mem-flashing partisi via Fastboot...", widget.NewActivity(), fp.window)
		progress.Show()

		go func() {
			fbClient := fp.deviceMgr.GetFastbootClient()
			out, err := fbClient.Run("flash", part, fp.selectedFilePath)

			fyne.Do(func() {
				progress.Hide()
				if err != nil {
					dialog.ShowError(fmt.Errorf("gagal flashing %s: %v\nOutput: %s", part, err, out), fp.window)
				} else {
					dialog.ShowInformation("Flashing Sukses!", fmt.Sprintf("Partisi %s berhasil di-flash!\n\nOutput:\n%s", part, out), fp.window)
				}
			})
		}()
	}, fp.window)
}

func (fp *FlasherPage) runBackup() {
	info := fp.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		dialog.ShowError(fmt.Errorf("perangkat harus terhubung dalam mode Android (ADB) untuk melakukan backup."), fp.window)
		return
	}

	part := fp.selectBackupPart.Selected
	progress := dialog.NewCustomWithoutButtons(fmt.Sprintf("Memproses backup partisi: %s...", part), widget.NewActivity(), fp.window)
	progress.Show()

	go func() {
		adbClient := fp.deviceMgr.GetADBClient()

		// 1. Check root
		rootCheck, _ := adbClient.Shell("su -c 'whoami' 2>/dev/null")
		if !strings.Contains(strings.ToLower(rootCheck), "root") {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("akses backup partisi langsung memerlukan izin root (HP harus di-root dan izin ADB root diberikan)."), fp.window)
			})
			return
		}

		// Create temp file path on device
		deviceTmp := fmt.Sprintf("/data/local/tmp/backup_%s.img", part)
		
		// 2. Dump partition to temp file
		ddCmd := fmt.Sprintf("su -c 'dd if=/dev/block/by-name/%s of=%s status=none'", part, deviceTmp)
		_, err := adbClient.Shell(ddCmd)
		if err != nil {
			fyne.Do(func() {
				progress.Hide()
				dialog.ShowError(fmt.Errorf("gagal melakukan dump partisi: %v", err), fp.window)
			})
			return
		}

		// 3. Prepare local path
		outputDir := GetLogOutputDir()
		_ = os.MkdirAll(outputDir, 0755)
		timestamp := time.Now().Format("20060102_150405")
		localFile := filepath.Join(outputDir, fmt.Sprintf("backup_%s_%s.img", part, timestamp))

		// 4. Pull to PC
		_, pullErr := adbClient.Run("pull", deviceTmp, localFile)
		
		// 5. Cleanup device temp file
		_, _ = adbClient.Shell(fmt.Sprintf("rm %s 2>/dev/null", deviceTmp))

		fyne.Do(func() {
			progress.Hide()
			if pullErr != nil {
				dialog.ShowError(fmt.Errorf("gagal memindahkan (pull) backup file ke PC: %v", pullErr), fp.window)
			} else {
				dialog.ShowInformation("Backup Berhasil!", fmt.Sprintf("Partisi %s berhasil dibackup!\n\nDisimpan ke PC:\n%s", part, localFile), fp.window)
			}
		})
	}()
}

func (fp *FlasherPage) runChecks() {
	fp.btnCheck.Disable()
	fp.lblSummary.SetText("Sedang memverifikasi kompatibilitas device secara mendalam...")
	fp.resultBox.Objects = nil
	fp.resultBox.Refresh()

	go func() {
		res, err := fp.val.RunChecks()
		fyne.Do(func() {
			fp.btnCheck.Enable()
			if err != nil {
				dialog.ShowError(err, fp.window)
				fp.lblSummary.SetText("Gagal validasi: Device tidak ditemukan atau ADB mati.")
				return
			}

			// Render results
			resultItems := container.NewVBox()
			for _, chk := range res.Checks {
				emoji := chk.Status.Emoji()

				var statusColor color.Color
				switch chk.Status {
				case validator.CheckPass:
					statusColor = colorSuccess
				case validator.CheckWarn:
					statusColor = colorWarning
				default:
					statusColor = colorError
				}

				nameLabel := widget.NewLabel(fmt.Sprintf("%s  %s", emoji, chk.Name))
				nameLabel.TextStyle = fyne.TextStyle{Bold: true}

				valLabel := NewNeoLabel(chk.Value)

				detailLabel := NewNeoLabel(chk.Detail)

				statusLine := canvas.NewRectangle(statusColor)
				statusLine.SetMinSize(fyne.NewSize(4, 0))

				itemRow := container.NewBorder(nil, nil, statusLine, nil, container.NewPadded(container.NewVBox(
					container.NewBorder(nil, nil, nameLabel, nil, container.NewHBox(layout.NewSpacer(), valLabel)),
					detailLabel,
				)))
				resultItems.Add(itemRow)
				resultItems.Add(NeoDivider())
			}

			fp.resultBox.Add(NewNeoCard("Hasil Pengecekan", "", resultItems))

			// Render overall summary
			summaryMsg := ""
			if res.AllPassed {
				if res.HasWarning {
					summaryMsg = "⚠️ PRE-FLASH COMPLETED DENGAN PERINGATAN!\nDevice Anda kompatibel tetapi ada catatan penting (misal: panel variasi LC0C/LC0D). Bacalah detail di bawah sebelum melanjutkan."
				} else {
					summaryMsg = "🎉 DEVICE ANDA 100% KOMPATIBEL & SIAP FLASH!\nSemua spesifikasi dan konfigurasi verified boot sesuai dengan persyaratan instalasi Custom Kernel."
				}
			} else {
				summaryMsg = "❌ DEVICE TIDAK KOMPATIBEL ATAU BUTUH PERSIAPAN!\nTerdeteksi parameter kritis yang gagal. JANGAN melakukan flashing Custom Kernel sebelum semua tanda merah diselesaikan terlebih dahulu demi mencegah hard-brick/bootloop."
			}
			
			fp.lblSummary.SetText(summaryMsg)
			fp.resultBox.Refresh()
		})
	}()
}
