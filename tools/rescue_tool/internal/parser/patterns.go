package parser

import "regexp"

// Severity represents the severity level of a detected pattern
type Severity int

const (
	INFO     Severity = iota
	WARNING
	CRITICAL
)

func (s Severity) String() string {
	switch s {
	case CRITICAL:
		return "CRITICAL"
	case WARNING:
		return "WARNING"
	default:
		return "INFO"
	}
}

// ErrorPattern defines a regex-based pattern with diagnosis info
type ErrorPattern struct {
	Pattern    *regexp.Regexp
	Severity   Severity
	Category   string // Grouping label (e.g. "Kernel", "WiFi", "KSU")
	Diagnosis  string // Human-readable diagnosis shown to user
	ActionHint string // Suggested next step
}

// Match holds the result of a single pattern match against a log line
type Match struct {
	LineNumber int
	LineText   string
	Pattern    *ErrorPattern
}

// AnalysisResult holds the full result of parsing a crash log
type AnalysisResult struct {
	TotalLines    int
	Matches       []Match
	TopIssues     []TopIssue
	SeverityCounts map[Severity]int
}

// TopIssue summarizes a detected problem for the UI
type TopIssue struct {
	Severity   Severity
	Category   string
	Diagnosis  string
	ActionHint string
	Count      int // How many lines matched this pattern
	FirstLine  int // First line number where this appeared
}

// DefaultPatterns returns the built-in error patterns for Epitaph Kernel crash logs
func DefaultPatterns() []*ErrorPattern {
	return []*ErrorPattern{
		{
			Pattern:    regexp.MustCompile(`(?i)Kernel panic`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "Kernel mengalami crash fatal (Kernel Panic)",
			ActionHint: "Lihat baris tepat di atas untuk menemukan trigger awal. Biasanya ada NULL pointer dereference atau out-of-bounds access.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)BUG: scheduling while atomic`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "Race condition terdeteksi — scheduling while atomic",
			ActionHint: "Kemungkinan SUSFS atau KSU hook menyebabkan deadlock. Coba flash kernel tanpa SUSFS.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)Unable to handle kernel (NULL )?pointer dereference`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "NULL pointer dereference — crash karena akses memori invalid",
			ActionHint: "Cek call stack di bawah baris ini untuk mengetahui fungsi mana yang crash.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)CONFIG_DEBUG_INFO_NONE`),
			Severity:   WARNING,
			Category:   "WiFi",
			Diagnosis:  "BPF/BTF metadata hilang — WiFi akan mati setelah flash via Fastboot",
			ActionHint: "Flash ulang kernel menggunakan KernelFlasher (bukan Fastboot) untuk mengembalikan modul vendor_dlkm.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)cfg80211.*(?:failed|error|unable)`),
			Severity:   CRITICAL,
			Category:   "WiFi",
			Diagnosis:  "Driver WiFi (cfg80211) gagal load",
			ActionHint: "Modul WiFi mismatch. Flash stock boot dulu, lalu flash Epitaph via KernelFlasher.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)module verification failed`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "KMI/vermagic mismatch — modul vendor ditolak kernel",
			ActionHint: "Kernel version tidak cocok dengan modul di /vendor_dlkm. Flash ulang ROM atau gunakan kernel yang matching.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)init: Service '([^']+)' .*(killed|died|failed)`),
			Severity:   WARNING,
			Category:   "Android",
			Diagnosis:  "Proses Android service mati — kemungkinan SELinux blocking",
			ActionHint: "Cek apakah SELinux dalam mode enforcing. Jika pakai KSU, pastikan patch SELinux sudah benar.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)avc:\s+denied`),
			Severity:   WARNING,
			Category:   "SELinux",
			Diagnosis:  "SELinux denial terdeteksi — ada akses yang diblokir",
			ActionHint: "Ini bisa menyebabkan service Android gagal start. Periksa policy SELinux.",
		},
		{
			Pattern:    regexp.MustCompile(`KSU:`),
			Severity:   INFO,
			Category:   "KernelSU",
			Diagnosis:  "KernelSU log entry",
			ActionHint: "Periksa detail KSU operation — biasanya informational.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)susfs`),
			Severity:   INFO,
			Category:   "SUSFS",
			Diagnosis:  "SUSFS activity terdeteksi",
			ActionHint: "Monitor log SUSFS untuk memastikan tidak ada conflict dengan modul lain.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)Out of memory`),
			Severity:   CRITICAL,
			Category:   "Memory",
			Diagnosis:  "Out of Memory (OOM) — sistem kehabisan RAM",
			ActionHint: "Kemungkinan memory leak di kernel module. Cek proses mana yang consume memory terbanyak.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)call trace:`),
			Severity:   WARNING,
			Category:   "Kernel",
			Diagnosis:  "Kernel call trace / stack dump",
			ActionHint: "Baris-baris setelah ini menunjukkan urutan function call saat error terjadi.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)RCU detected stall`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "RCU stall terdeteksi — CPU terjebak di critical section terlalu lama",
			ActionHint: "Ini menandakan deadlock atau infinite loop di kernel. Flash stock boot untuk recover.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)Watchdog bark`),
			Severity:   CRITICAL,
			Category:   "Kernel",
			Diagnosis:  "Watchdog bark — sistem hang dan watchdog timer expired",
			ActionHint: "Device hang total. Flash stock boot dan pull log untuk debugging.",
		},
		{
			Pattern:    regexp.MustCompile(`(?i)thermal emergency`),
			Severity:   WARNING,
			Category:   "Hardware",
			Diagnosis:  "Thermal emergency — device terlalu panas",
			ActionHint: "Cek apakah ada proses yang menyebabkan CPU terus-menerus 100%.",
		},
	}
}
