#!/system/bin/sh
# ==============================================================================
#  Epitaph Kernel Optimization & Reliability Tuner (Thermal & Charging Aware)
#  Designed by Naidrahiqa & Antigravity AI
#  Epitaph Kernel — Redmi 12 (fire) — GKI 6.6
# ==============================================================================
# File ini diletakkan di /data/adb/service.d/epitaph_tuner.sh oleh AnyKernel3
# Berjalan setiap boot via KernelSU/Magisk service.d atau runtime secara manual
# ==============================================================================

sleep 5

LOG_FILE="/data/local/tmp/epitaph_tuner.log"
STATUS_FILE="/data/adb/epitaph/status"
MODE_FILE="/data/adb/epitaph/mode"
APPLY_FILE="/data/adb/epitaph/apply"

# Direktori logging terdedikasi
mkdir -p /data/local/tmp 2>/dev/null
mkdir -p /data/adb/epitaph 2>/dev/null
mkdir -p /data/epitaph 2>/dev/null

chmod 644 "$LOG_FILE" 2>/dev/null

log_msg() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

log_thermal() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "/data/epitaph/thermal.log"
}

log_charging() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "/data/epitaph/charging.log"
}

# Helper: menulis ke sysfs/procfs secara aman tanpa warning
write_value() {
  local val="$1"
  local target="$2"
  if [ -e "$target" ]; then
    { echo "$val" > "$target"; } 2>/dev/null
  fi
}

# Helper: menyalin konten berkas secara aman
copy_value() {
  local src="$1"
  local target="$2"
  if [ -f "$src" ] && [ -e "$target" ]; then
    { cat "$src" > "$target"; } 2>/dev/null
  fi
}

# ──────────────────────────────────────────────────────────────────────────────
# DAEMON SUBROUTINES (Mencegah Eksekusi Ganda via Daemon Flag)
# ──────────────────────────────────────────────────────────────────────────────

# 1. Thermal-Aware Daemon Loop
thermal_daemon() {
  log_thermal "=== THERMAL DAEMON STARTED ==="
  local last_state=""
  
  # Identifikasi CPU thermal zone dinamis untuk Helio G88 (MT6769)
  local cpu_zone=""
  for tz in /sys/class/thermal/thermal_zone*; do
    if [ -f "$tz/type" ]; then
      local type=$(cat "$tz/type" | tr '[:upper:]' '[:lower:]')
      if echo "$type" | grep -qE "cpu|soc|mtktscpu"; then
        cpu_zone="$tz"
        break
      fi
    fi
  done
  
  [ -z "$cpu_zone" ] && cpu_zone="/sys/class/thermal/thermal_zone0"
  log_thermal "Thermal zone terpilih: $cpu_zone ($(cat $cpu_zone/type 2>/dev/null || echo 'unknown'))"
  
  while true; do
    local temp_raw=$(cat "$cpu_zone/temp" 2>/dev/null || echo "0")
    local temp=$((temp_raw / 1000))
    local current_state="WARM"
    
    if [ "$temp" -lt 40 ]; then
      current_state="COOL"
    elif [ "$temp" -gt 55 ]; then
      current_state="HOT"
    else
      current_state="WARM"
    fi
    
    # Deteksi transisi status thermal
    if [ "$current_state" != "$last_state" ]; then
      log_thermal "Transisi Suhu: ${last_state:-NONE} -> ${current_state} (${temp}°C)"
      last_state="$current_state"
      echo "$current_state" > "/data/adb/epitaph/thermal_state" 2>/dev/null
      
      # Terapkan pembatasan thermal governor & clock secara dinamis
      apply_thermal_tuning "$current_state"
    fi
    
    # Jalankan pengecekan memory pressure LMKD dinamis di loop yang sama
    tune_lmkd
    
    sleep 10
  done
}

# 2. Charging-State Boost Daemon Loop
charging_daemon() {
  log_charging "=== CHARGING DAEMON STARTED ==="
  local last_status=""
  
  while true; do
    local status=$(cat /sys/class/power_supply/battery/status 2>/dev/null | tr -d ' \r\n')
    [ -z "$status" ] && status="Discharging"
    
    if [ "$status" != "$last_status" ]; then
      log_charging "Transisi Daya: ${last_status:-NONE} -> ${status}"
      last_status="$status"
      
      local therm_state=$(cat "/data/adb/epitaph/thermal_state" 2>/dev/null || echo "WARM")
      
      if [ "$status" = "Charging" ] || [ "$status" = "Full" ]; then
        if [ "$therm_state" = "HOT" ]; then
          log_charging "🚨 Suhu terlalu panas (${therm_state}). Boost pengisian daya dilewati."
          revert_charging_boost
        else
          apply_charging_boost
        fi
      else
        revert_charging_boost
      fi
    fi
    
    # Proteksi real-time jika perangkat tiba-tiba memanas saat dicas
    if [ "$status" = "Charging" ] || [ "$status" = "Full" ]; then
      local therm_state=$(cat "/data/adb/epitaph/thermal_state" 2>/dev/null || echo "WARM")
      local boost_active=$(cat "/data/adb/epitaph/charging_boost_active" 2>/dev/null || echo "false")
      
      if [ "$therm_state" = "HOT" ] && [ "$boost_active" = "true" ]; then
        log_charging "🚨 Perangkat memanas saat dicas! Mencabut boost pengisian daya secara darurat."
        revert_charging_boost
      elif [ "$therm_state" != "HOT" ] && [ "$boost_active" = "false" ]; then
        log_charging "⚡ Suhu stabil kembali. Menerapkan ulang boost pengisian daya."
        apply_charging_boost
      fi
    fi
    
    sleep 10
  done
}

# Tangkap flag daemon sebelum inisialisasi utama untuk menghindari instansiasi ganda
if [ "$1" = "--thermal-daemon" ]; then
  thermal_daemon
  exit 0
elif [ "$1" = "--charging-daemon" ]; then
  charging_daemon
  exit 0
fi

# ──────────────────────────────────────────────────────────────────────────────
# 3-MODE POWER PROFILE DEFINITIONS (Helio G88 Optimizations)
# ──────────────────────────────────────────────────────────────────────────────

battery() {
  log_msg "Tuning profile: battery"
  
  # CPU Frequency Caps (A55 Little Core & A75 Big Core)
  # LITTLE Cluster (policy0) batasi max ke 1.38GHz
  write_value 500000 /sys/devices/system/cpu/cpufreq/policy0/scaling_min_freq
  write_value 1380000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
  
  # BIG Cluster (policy6) batasi max ke 1.38GHz
  write_value 900000 /sys/devices/system/cpu/cpufreq/policy6/scaling_min_freq
  write_value 1380000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
  
  # Schedutil Conservativeness
  for policy in /sys/devices/system/cpu/cpufreq/policy*; do
    if [ -f "$policy/scaling_governor" ] && [ "$(cat $policy/scaling_governor)" = "schedutil" ]; then
      write_value 2000 "$policy/schedutil/up_rate_limit_us"
      write_value 1000 "$policy/schedutil/down_rate_limit_us"
      write_value 95 "$policy/schedutil/hispeed_load"
      
      # Tentukan frekuensi hispeed hemat
      p_num="${policy##*policy}"
      if [ "$p_num" -eq 6 ]; then
        write_value 1150000 "$policy/schedutil/hispeed_freq"
        write_value 0 "$policy/schedutil/epitaph_boost_factor"
        write_value 95 "$policy/schedutil/epitaph_boost_threshold"
      else
        write_value 1100000 "$policy/schedutil/hispeed_freq"
        write_value 0 "$policy/schedutil/epitaph_boost_factor"
        write_value 95 "$policy/schedutil/epitaph_boost_threshold"
      fi
    fi
  done
  
  # CPU Uclamp - Maksimal Penghematan Baterai
  write_value 0 /dev/cpuctl/cpu.uclamp.min
  write_value 0 /dev/cpuctl/top-app/cpu.uclamp.min
  write_value 0 /dev/cpuctl/foreground/cpu.uclamp.min
  write_value 0 /dev/cpuctl/background/cpu.uclamp.min
  write_value 0 /dev/cpuctl/system-background/cpu.uclamp.min
  
  # GPU Mali & GED Power Saving Settings
  write_value 0 /sys/kernel/ged/hal/gpu_boost
  write_value 0 /sys/module/ged/parameters/boost_gpu_enable
  for mali_dir in /sys/class/misc/mali0/device /sys/devices/platform/*.mali; do
    if [ -d "$mali_dir" ]; then
      write_value "coarse_demand" "$mali_dir/power_policy"
    fi
  done
  
  # Virtual Memory (Swappiness Rendah, Flush Agresif)
  write_value 160 /proc/sys/vm/swappiness
  write_value 20 /proc/sys/vm/dirty_ratio
  write_value 5 /proc/sys/vm/dirty_background_ratio
  write_value 300 /proc/sys/vm/dirty_writeback_centisecs
  write_value 2000 /proc/sys/vm/dirty_expire_centisecs
  
  # EAS Scheduler Latency (Batasi Siklus Bangun CPU)
  write_value 24000000 /proc/sys/kernel/sched_latency_ns
  write_value 4000000 /proc/sys/kernel/sched_min_granularity_ns
  write_value 6000000 /proc/sys/kernel/sched_wakeup_granularity_ns
  
  # Cpuset (Batasi Latar Belakang ke Little Core)
  write_value "0-5" /dev/cpuset/background/cpus
  write_value "0-5" /dev/cpuset/system-background/cpus
  write_value "0-5" /dev/cpuset/restricted/cpus
}

balanced() {
  log_msg "Tuning profile: balanced"
  
  # CPU Frequency (Full Range untuk Little & Big)
  write_value 500000 /sys/devices/system/cpu/cpufreq/policy0/scaling_min_freq
  write_value 1800000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
  
  write_value 900000 /sys/devices/system/cpu/cpufreq/policy6/scaling_min_freq
  write_value 2000000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
  
  # Schedutil Balanced Values
  for policy in /sys/devices/system/cpu/cpufreq/policy*; do
    if [ -f "$policy/scaling_governor" ] && [ "$(cat $policy/scaling_governor)" = "schedutil" ]; then
      write_value 500 "$policy/schedutil/up_rate_limit_us"
      write_value 10000 "$policy/schedutil/down_rate_limit_us"
      write_value 85 "$policy/schedutil/hispeed_load"
      
      p_num="${policy##*policy}"
      if [ "$p_num" -eq 6 ]; then
        write_value 1500000 "$policy/schedutil/hispeed_freq"
        write_value 15 "$policy/schedutil/epitaph_boost_factor"
        write_value 60 "$policy/schedutil/epitaph_boost_threshold"
      else
        write_value 1380000 "$policy/schedutil/hispeed_freq"
        write_value 5 "$policy/schedutil/epitaph_boost_factor"
        write_value 80 "$policy/schedutil/epitaph_boost_threshold"
      fi
    fi
  done
  
  # CPU Uclamp - Responsif tapi Ramah Baterai
  write_value 0 /dev/cpuctl/cpu.uclamp.min
  write_value 64 /dev/cpuctl/top-app/cpu.uclamp.min
  write_value 16 /dev/cpuctl/foreground/cpu.uclamp.min
  write_value 0 /dev/cpuctl/background/cpu.uclamp.min
  write_value 0 /dev/cpuctl/system-background/cpu.uclamp.min
  
  # GPU Mali & GED Balanced Settings
  write_value 0 /sys/kernel/ged/hal/gpu_boost
  write_value 1 /sys/module/ged/parameters/boost_gpu_enable
  for mali_dir in /sys/class/misc/mali0/device /sys/devices/platform/*.mali; do
    if [ -d "$mali_dir" ]; then
      write_value "dynamic" "$mali_dir/power_policy"
    fi
  done
  
  # Virtual Memory Balanced
  write_value 180 /proc/sys/vm/swappiness
  write_value 15 /proc/sys/vm/dirty_ratio
  write_value 3 /proc/sys/vm/dirty_background_ratio
  write_value 150 /proc/sys/vm/dirty_writeback_centisecs
  write_value 1000 /proc/sys/vm/dirty_expire_centisecs
  
  # EAS Scheduler Latency Balanced
  write_value 16000000 /proc/sys/kernel/sched_latency_ns
  write_value 3000000 /proc/sys/kernel/sched_min_granularity_ns
  write_value 4000000 /proc/sys/kernel/sched_wakeup_granularity_ns
  
  # Cpuset Balanced
  write_value "0-5" /dev/cpuset/background/cpus
  write_value "0-5" /dev/cpuset/system-background/cpus
  write_value "0-5" /dev/cpuset/restricted/cpus
}

performance() {
  log_msg "Tuning profile: performance"
  
  # CPU Frequency (Buka Limit Bawah Core untuk Menghilangkan Stutter)
  write_value 700000 /sys/devices/system/cpu/cpufreq/policy0/scaling_min_freq
  write_value 1800000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
  
  write_value 1150000 /sys/devices/system/cpu/cpufreq/policy6/scaling_min_freq
  write_value 2000000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
  
  # Schedutil Performance Values
  for policy in /sys/devices/system/cpu/cpufreq/policy*; do
    if [ -f "$policy/scaling_governor" ] && [ "$(cat $policy/scaling_governor)" = "schedutil" ]; then
      write_value 100 "$policy/schedutil/up_rate_limit_us"
      write_value 40000 "$policy/schedutil/down_rate_limit_us"
      write_value 75 "$policy/schedutil/hispeed_load"
      
      p_num="${policy##*policy}"
      if [ "$p_num" -eq 6 ]; then
        write_value 1850000 "$policy/schedutil/hispeed_freq"
        write_value 40 "$policy/schedutil/epitaph_boost_factor"
        write_value 30 "$policy/schedutil/epitaph_boost_threshold"
      else
        write_value 1600000 "$policy/schedutil/hispeed_freq"
        write_value 15 "$policy/schedutil/epitaph_boost_factor"
        write_value 60 "$policy/schedutil/epitaph_boost_threshold"
      fi
    fi
  done
  
  # CPU Uclamp - Responsivitas Maksimal untuk UI & Game (Tidak 1024 agar deepsleep aktif)
  write_value 0 /dev/cpuctl/cpu.uclamp.min
  write_value 180 /dev/cpuctl/top-app/cpu.uclamp.min
  write_value 64 /dev/cpuctl/foreground/cpu.uclamp.min
  write_value 0 /dev/cpuctl/background/cpu.uclamp.min
  write_value 0 /dev/cpuctl/system-background/cpu.uclamp.min
  
  # GPU Mali & GED High Boost Settings
  write_value 1 /sys/kernel/ged/hal/gpu_boost
  write_value 1 /sys/module/ged/parameters/boost_gpu_enable
  for mali_dir in /sys/class/misc/mali0/device /sys/devices/platform/*.mali; do
    if [ -d "$mali_dir" ]; then
      write_value "always_on" "$mali_dir/power_policy"
      if [ -f "$mali_dir/dvfs_max_freq" ]; then
        if [ -f "$mali_dir/dvfs_max_freq_khz" ]; then
          copy_value "$mali_dir/dvfs_max_freq_khz" "$mali_dir/dvfs_max_freq"
        elif [ -f "$mali_dir/max_clock" ]; then
          copy_value "$mali_dir/max_clock" "$mali_dir/dvfs_max_freq"
        fi
      fi
    fi
  done
  
  # Virtual Memory (Swappiness Tinggi untuk ZRAM, Sinkronisasi I/O Sangat Cepat)
  write_value 200 /proc/sys/vm/swappiness
  write_value 10 /proc/sys/vm/dirty_ratio
  write_value 2 /proc/sys/vm/dirty_background_ratio
  write_value 100 /proc/sys/vm/dirty_writeback_centisecs
  write_value 500 /proc/sys/vm/dirty_expire_centisecs
  
  # EAS Scheduler Latency Rendah (Menghindari Frame Drop)
  write_value 10000000 /proc/sys/kernel/sched_latency_ns
  write_value 1500000 /proc/sys/kernel/sched_min_granularity_ns
  write_value 2000000 /proc/sys/kernel/sched_wakeup_granularity_ns
  
  # Cpuset (Buka Semua Core untuk Background Services)
  write_value "0-7" /dev/cpuset/background/cpus
  write_value "0-7" /dev/cpuset/system-background/cpus
  write_value "0-7" /dev/cpuset/restricted/cpus
}

# ──────────────────────────────────────────────────────────────────────────────
# THERMAL-AWARE DYNAMIC SCALING
# ──────────────────────────────────────────────────────────────────────────────

apply_thermal_tuning() {
  local state="$1"
  log_thermal "Menerapkan limit thermal state: $state"
  
  case "$state" in
    COOL)
      # Kembalikan clock maksimal dan kurangi up_rate_limit untuk akselerasi instan
      write_value 1800000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
      write_value 2000000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
      
      for policy in /sys/devices/system/cpu/cpufreq/policy*; do
        if [ -f "$policy/schedutil/up_rate_limit_us" ]; then
          if [ "$MODE" = "performance" ]; then
            write_value 100 "$policy/schedutil/up_rate_limit_us"
          else
            write_value 500 "$policy/schedutil/up_rate_limit_us"
          fi
        fi
      done
      ;;
      
    WARM)
      # Profil standar harian
      write_value 1800000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
      write_value 2000000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
      
      for policy in /sys/devices/system/cpu/cpufreq/policy*; do
        if [ -f "$policy/schedutil/up_rate_limit_us" ]; then
          if [ "$MODE" = "performance" ]; then
            write_value 100 "$policy/schedutil/up_rate_limit_us"
          elif [ "$MODE" = "battery" ]; then
            write_value 2000 "$policy/schedutil/up_rate_limit_us"
          else
            write_value 500 "$policy/schedutil/up_rate_limit_us"
          fi
        fi
      done
      ;;
      
    HOT)
      # Throttling Aktif: Pangkas clock atas untuk mencegah hardware degradation
      log_thermal "🚨 THROTILING AKTIF: Batasi frekuensi maksimal core!"
      write_value 1500000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
      write_value 1600000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
      
      # Naikkan rate_limit untuk menstabilkan voltase/suhu
      for policy in /sys/devices/system/cpu/cpufreq/policy*; do
        write_value 2000 "$policy/schedutil/up_rate_limit_us"
      done
      
      # Turunkan swappiness untuk mengurangi beban kerja CPU dari kompresi ZRAM intensif
      write_value 100 /proc/sys/vm/swappiness
      ;;
  esac
}

# ──────────────────────────────────────────────────────────────────────────────
# CHARGING-STATE BOOST SCALING
# ──────────────────────────────────────────────────────────────────────────────

apply_charging_boost() {
  echo "true" > "/data/adb/epitaph/charging_boost_active" 2>/dev/null
  log_charging "⚡ Mengaktifkan OC Charging Boost!"
  
  # Buka frekuensi maksimal dan percepat lompatan frekuensi
  write_value 1800000 /sys/devices/system/cpu/cpufreq/policy0/scaling_max_freq
  write_value 2000000 /sys/devices/system/cpu/cpufreq/policy6/scaling_max_freq
  
  for policy in /sys/devices/system/cpu/cpufreq/policy*; do
    write_value 100 "$policy/schedutil/up_rate_limit_us"
    write_value 40000 "$policy/schedutil/down_rate_limit_us"
  done
  
  # Dorong clock GPU agar tetap stabil saat beban berat serentak
  write_value 1 /sys/kernel/ged/hal/gpu_boost
}

revert_charging_boost() {
  echo "false" > "/data/adb/epitaph/charging_boost_active" 2>/dev/null
  log_charging "🔋 Menormalkan profil pengisian daya (kembali ke profil user: $MODE)"
  
  # Re-apply mode aktif saat ini untuk menormalkan governor
  apply_profile "$MODE"
}

# ──────────────────────────────────────────────────────────────────────────────
# DYNAMIC LOW MEMORY KILLER (LMKD) TUNING (4-6GB RAM Variants)
# ──────────────────────────────────────────────────────────────────────────────

tune_lmkd() {
  local mem_avail_kb=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
  local mem_avail_mb=$((mem_avail_kb / 1024))
  
  local pressure_tier="comfortable"
  if [ "$mem_avail_mb" -lt 400 ]; then
    pressure_tier="tight"
  elif [ "$mem_avail_mb" -lt 800 ]; then
    pressure_tier="moderate"
  else
    pressure_tier="comfortable"
  fi
  
  local minfree_pages=""
  
  case "$MODE" in
    performance)
      # Simpan aplikasi sebanyak mungkin di latar belakang untuk pengalaman multitasking yang mulus
      case "$pressure_tier" in
        comfortable)
          # Sangat santai: 48MB, 64MB, 80MB, 96MB, 128MB, 180MB
          minfree_pages="12288,16384,20480,24576,32768,46080"
          ;;
        moderate)
          # Sedang: 72MB, 90MB, 108MB, 126MB, 180MB, 240MB
          minfree_pages="18432,23040,27648,32256,46080,61440"
          ;;
        tight)
          # Agresif (menghindari OOM freeze): 90MB, 110MB, 130MB, 150MB, 220MB, 320MB
          minfree_pages="23040,28160,33280,38400,56320,81920"
          ;;
      esac
      ;;
    battery)
      # Agresif mematikan aplikasi latar belakang untuk menghemat daya konsumsi idle
      case "$pressure_tier" in
        comfortable)
          # Cukup ketat: 72MB, 90MB, 108MB, 126MB, 200MB, 280MB
          minfree_pages="18432,23040,27648,32256,51200,71680"
          ;;
        moderate)
          # Ketat: 90MB, 110MB, 130MB, 160MB, 240MB, 350MB
          minfree_pages="23040,28160,33280,40960,61440,89600"
          ;;
        tight)
          # Sangat agresif: 120MB, 140MB, 180MB, 220MB, 320MB, 450MB
          minfree_pages="30720,35840,46080,56320,81920,115200"
          ;;
      esac
      ;;
    balanced|*)
      # Profil seimbang standar untuk penggunaan harian
      case "$pressure_tier" in
        comfortable)
          # 60MB, 80MB, 100MB, 120MB, 160MB, 220MB
          minfree_pages="15360,20480,25600,30720,40960,56320"
          ;;
        moderate)
          # 72MB, 90MB, 108MB, 135MB, 200MB, 300MB
          minfree_pages="18432,23040,27648,34560,51200,76800"
          ;;
        tight)
          # 96MB, 120MB, 150MB, 180MB, 270MB, 380MB
          minfree_pages="24576,30720,38400,46080,69120,97280"
          ;;
      esac
      ;;
  esac
  
  # Terapkan ke parameter driver lowmemorykiller kernel jika aktif
  if [ -e "/sys/module/lowmemorykiller/parameters/minfree" ]; then
    write_value "$minfree_pages" /sys/module/lowmemorykiller/parameters/minfree
  fi
  
  # Setel Android LMKD properti level minfree dinamis
  setprop sys.lmk.minfree_levels "$minfree_pages" 2>/dev/null
}

# ──────────────────────────────────────────────────────────────────────────────
# SYSTEM INITIALIZATION & MAIN EXECUTION
# ──────────────────────────────────────────────────────────────────────────────

apply_profile() {
  local target_mode="$1"
  log_msg "Menerapkan profil daya utama: $target_mode"
  case "$target_mode" in
    battery)
      battery
      ;;
    performance)
      performance
      ;;
    balanced|*)
      balanced
      ;;
  esac
  
  # Sinkronisasi status persistensi profil
  echo "$target_mode" > "$MODE_FILE" 2>/dev/null
}

log_msg "=== EPITAPH TUNER TUNING IN PROGRESS ==="

# 1. WIFI MODULE LOADER & RECOVERY (MEMPERTAHANKAN COMPATIBILITY DARI PENYEDOT LAMA)
log_msg "Langkah 1: Menjalankan WiFi Module Loader..."
CFG_LOADED=false
if lsmod | grep -q cfg80211; then
  log_msg "cfg80211 sudah termuat"
  CFG_LOADED=true
else
  for search_dir in /vendor/lib/modules /vendor_dlkm/lib/modules /data/adb/wifi_fix; do
    if [ -f "$search_dir/cfg80211.ko" ]; then
      insmod "$search_dir/rfkill.ko" 2>/dev/null
      insmod "$search_dir/libarc4.ko" 2>/dev/null
      insmod "$search_dir/cfg80211.ko" 2>/dev/null
      if lsmod | grep -q cfg80211; then
        CFG_LOADED=true
        insmod "$search_dir/mac80211.ko" 2>/dev/null
        break
      fi
    fi
  done
fi

WLAN_LOADED=false
if lsmod | grep -qE "wlan_drv_gen4m"; then
  log_msg "Vendor WiFi driver sudah termuat"
  WLAN_LOADED=true
elif [ "$CFG_LOADED" = "true" ]; then
  for wlan_dir in /vendor/lib/modules /vendor_dlkm/lib/modules; do
    for wlan_file in wlan_drv_gen4m_6768.ko wlan_drv_gen4m.ko; do
      if [ -f "$wlan_dir/$wlan_file" ]; then
        insmod "$wlan_dir/$wlan_file" 2>/dev/null
        if lsmod | grep -qE "wlan_drv_gen4m"; then
          WLAN_LOADED=true
          break 2
        fi
      fi
    done
  done
fi

# 2. BACALAH PROP SISTEM DAN MODE PERSISTEN
PROP_MODE=$(getprop epitaph.profile 2>/dev/null | tr -d ' \r\n')
if [ -n "$PROP_MODE" ]; then
  MODE="$PROP_MODE"
else
  MODE=$(cat "$MODE_FILE" 2>/dev/null | tr -d ' \r\n')
fi

[ -z "$MODE" ] && MODE="balanced"
if [ "$MODE" != "performance" ] && [ "$MODE" != "balanced" ] && [ "$MODE" != "battery" ]; then
  MODE="balanced"
fi

# Terapkan profil utama
apply_profile "$MODE"

# 3. ZRAM & STORAGE OPTIMIZATIONS (STATIC INITS)
log_msg "Langkah 3: Menginisialisasi VM, ZRAM, dan Block IO Scheduler..."
write_value 100 /proc/sys/vm/vfs_cache_pressure

# ZRAM 6GB Setup
ZRAM_SIZE=6442450944
if [ "$(cat /sys/block/zram0/disksize 2>/dev/null || echo 0)" != "$ZRAM_SIZE" ]; then
  swapoff /dev/block/zram0 2>/dev/null || true
  write_value 1 /sys/block/zram0/reset
  if grep -q "zstd" /sys/block/zram0/comp_algorithm 2>/dev/null; then
    write_value "zstd" /sys/block/zram0/comp_algorithm
  else
    write_value "lz4" /sys/block/zram0/comp_algorithm
  fi
  write_value "$ZRAM_SIZE" /sys/block/zram0/disksize
  write_value 2 /sys/block/zram0/max_comp_streams
  mkswap /dev/block/zram0 2>/dev/null || true
  swapon /dev/block/zram0 -p 32767 2>/dev/null || true
fi

# Antrean Blok Penyimpanan eMMC 5.1 & Scheduler
for queue in /sys/block/*/queue; do
  if [ -d "$queue" ]; then
    write_value 512 "$queue/read_ahead_kb"
    write_value 0 "$queue/add_random"
    write_value 0 "$queue/rotational"
    write_value 0 "$queue/iostats"
    write_value 2 "$queue/rq_affinity"
    write_value 1 "$queue/nomerges"
    if [ -f "$queue/scheduler" ]; then
      if grep -q "kyber" "$queue/scheduler" 2>/dev/null && [ "$MODE" = "performance" ]; then
        write_value "kyber" "$queue/scheduler"
      else
        write_value "mq-deadline" "$queue/scheduler"
      fi
    fi
  fi
done

# Optimasi TCP BBR
write_value "bbr" /proc/sys/net/ipv4/tcp_congestion_control
write_value "fq" /proc/sys/net/core/default_qdisc
write_value 3 /proc/sys/net/ipv4/tcp_fastopen
write_value 0 /proc/sys/net/ipv4/tcp_slow_start_after_idle

# Membuat skrip apply instan agar runtime manual berjalan mulus
cat << 'EOF' > "$APPLY_FILE"
#!/system/bin/sh
# Trigger re-apply Epitaph Schedutil profile real-time tanpa reboot
/system/bin/sh /data/adb/service.d/epitaph_tuner.sh
EOF
chmod 755 "$APPLY_FILE" 2>/dev/null

# 4. MEMULAI BACKGROUND MONITORING SECARA AMAN (Cegah Duplikasi Daemon)
log_msg "Langkah 4: Melakukan sterilisasi dan memicu proses background daemons..."

# Hentikan semua daemon lama yang sedang berjalan
pkill -f "epitaph_tuner.sh --thermal-daemon" || true
pkill -f "epitaph_tuner.sh --charging-daemon" || true

# Jalankan daemon baru di latar belakang secara bersih
/system/bin/sh /data/adb/service.d/epitaph_tuner.sh --thermal-daemon >/dev/null 2>&1 &
/system/bin/sh /data/adb/service.d/epitaph_tuner.sh --charging-daemon >/dev/null 2>&1 &

# Buat berkas status untuk info user / Kernel Manager
echo "active_profile: $MODE" > "$STATUS_FILE"
echo "wifi_status: cfg=$CFG_LOADED, vendor=$WLAN_LOADED" >> "$STATUS_FILE"
echo "thermal_monitor: active" >> "$STATUS_FILE"
echo "charging_boost: ready" >> "$STATUS_FILE"
echo "last_applied: $(date)" >> "$STATUS_FILE"

log_msg "=== EPITAPH TUNER INITIALIZATION COMPLETED SUCCESSFULLY ==="
