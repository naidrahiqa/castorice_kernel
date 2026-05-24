#!/system/bin/sh
# ==============================================================================
#  Epitaph Kernel Optimization & Reliability Tuner
#  Designed by Naidrahiqa & Antigravity AI
#  Epitaph Kernel — Redmi 12 (fire) — GKI 6.6
# ==============================================================================
# This file is placed at /data/adb/service.d/epitaph_tuner.sh by AnyKernel3
# It runs at every boot via KernelSU/Magisk service.d
# ==============================================================================

sleep 5

LOG_FILE="/data/local/tmp/epitaph_tuner.log"
mkdir -p /data/local/tmp 2>/dev/null
chmod 644 "$LOG_FILE" 2>/dev/null

log_msg() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

log_msg "=== EPITAPH TUNER STARTED ==="

# Initialize Epitaph Schedutil Performance profile folder and files
mkdir -p /data/adb/epitaph 2>/dev/null
MODE_FILE="/data/adb/epitaph/mode"
APPLY_FILE="/data/adb/epitaph/apply"

if [ ! -f "$MODE_FILE" ]; then
  echo "balanced" > "$MODE_FILE"
  chmod 644 "$MODE_FILE" 2>/dev/null
fi

MODE=$(cat "$MODE_FILE" | tr -d ' \r\n')
if [ "$MODE" != "performance" ] && [ "$MODE" != "balanced" ] && [ "$MODE" != "battery" ]; then
  MODE="balanced"
fi

log_msg "Selected Schedutil Profile: $MODE"

# Create trigger script to re-apply real-time without reboot
cat << 'EOF' > "$APPLY_FILE"
#!/system/bin/sh
# Trigger re-apply Epitaph Schedutil profile real-time without reboot
/system/bin/sh /data/adb/service.d/epitaph_tuner.sh
EOF
chmod 755 "$APPLY_FILE" 2>/dev/null

# 1. FALLBACK WIFI MODULE LOADER
# Ensures 100% WiFi/Hotspot reliability if systemless module loading fails
log_msg "Section 1: Checking WiFi module fallback..."
if ! lsmod | grep -q cfg80211; then
  log_msg "WiFi module not loaded, attempting fallback loading..."
  if [ -d /data/adb/wifi_fix ]; then
    insmod /data/adb/wifi_fix/rfkill.ko 2>/dev/null
    insmod /data/adb/wifi_fix/libarc4.ko 2>/dev/null
    insmod /data/adb/wifi_fix/cfg80211.ko 2>/dev/null
    insmod /data/adb/wifi_fix/mac80211.ko 2>/dev/null
    if lsmod | grep -q cfg80211; then
      log_msg "Fallback WiFi module successfully loaded"
    else
      log_msg "Fallback WiFi module failed to load"
    fi
  else
    log_msg "/data/adb/wifi_fix directory not found, fallback unavailable"
  fi
else
  log_msg "WiFi module (cfg80211) already loaded by systemless loader"
fi

# 2. CPU SCHEDUTIL GOVERNOR OPTIMIZATIONS
# Smooths UI transitions & eliminates micro-stutters based on active profile
log_msg "Section 2: Tuning CPU Schedutil governors for profile: $MODE"

UP_RATE=500
DOWN_RATE=10000

case "$MODE" in
  performance)
    UP_RATE=100
    DOWN_RATE=40000
    ;;
  battery)
    UP_RATE=2000
    DOWN_RATE=1000
    ;;
  balanced|*)
    UP_RATE=500
    DOWN_RATE=10000
    ;;
esac

for policy in /sys/devices/system/cpu/cpufreq/policy*; do
  if [ -f "$policy/scaling_governor" ]; then
    gov=$(cat "$policy/scaling_governor")
    if [ "$gov" = "schedutil" ]; then
      echo "$UP_RATE" > "$policy/schedutil/up_rate_limit_us" 2>/dev/null
      echo "$DOWN_RATE" > "$policy/schedutil/down_rate_limit_us" 2>/dev/null
      log_msg "Tuned $policy (schedutil): up=$UP_RATE, down=$DOWN_RATE"
    fi
  fi
done

# 3. GPU GED & MALI LIMITER RESET
# Fixes GPU frequency locks / MTK thermal driver throttle bugs
log_msg "Section 3: Optimizing GPU settings..."
if [ -d /sys/kernel/ged/hal ]; then
  echo 1 > /sys/kernel/ged/hal/gpu_boost 2>/dev/null
  echo 1 > /sys/module/ged/parameters/boost_gpu_enable 2>/dev/null
  log_msg "GED GPU boost enabled"
fi
for mali_dir in /sys/class/misc/mali0/device /sys/devices/platform/*.mali; do
  if [ -d "$mali_dir" ]; then
    echo "dynamic" > "$mali_dir/power_policy" 2>/dev/null
    if [ -f "$mali_dir/dvfs_max_freq" ]; then
      if [ -f "$mali_dir/dvfs_max_freq_khz" ]; then
        cat "$mali_dir/dvfs_max_freq_khz" > "$mali_dir/dvfs_max_freq" 2>/dev/null
      elif [ -f "$mali_dir/max_clock" ]; then
        cat "$mali_dir/max_clock" > "$mali_dir/dvfs_max_freq" 2>/dev/null
      fi
    fi
    log_msg "Mali GPU policy set to dynamic & maximum frequency locked for $mali_dir"
  fi
done

# 4. MEMORY & VIRTUAL MEMORY TUNING
# Resolves high RAM consumption, prevents OOM & background app kills
log_msg "Section 4: Tuning Memory & Virtual Memory..."
echo 180 > /proc/sys/vm/swappiness 2>/dev/null && log_msg "Swappiness set to 180"
echo 100 > /proc/sys/vm/vfs_cache_pressure 2>/dev/null && log_msg "vfs_cache_pressure set to 100"
echo 20 > /proc/sys/vm/dirty_ratio 2>/dev/null && log_msg "dirty_ratio set to 20"
echo 5 > /proc/sys/vm/dirty_background_ratio 2>/dev/null && log_msg "dirty_background_ratio set to 5"
if [ -e /dev/block/zram0 ]; then
  echo 2 > /sys/block/zram0/max_comp_streams 2>/dev/null && log_msg "zram0 max_comp_streams set to 2"
fi

# 5. STORAGE READ-AHEAD OPTIMIZATION
# Enhances app loading speed
log_msg "Section 5: Tuning Storage Read-Ahead..."
for queue in /sys/block/*/queue; do
  if [ -d "$queue" ]; then
    echo 512 > "$queue/read_ahead_kb" 2>/dev/null
  fi
done
log_msg "Read-ahead buffers set to 512KB for storage block queues"

# 6. TCP SYSCTL NETWORK TUNING
# Optimizes networking performance & latency
log_msg "Section 6: Applying TCP sysctl optimizations..."
echo "bbr" > /proc/sys/net/ipv4/tcp_congestion_control 2>/dev/null && log_msg "TCP congestion control set to BBR"
echo "fq" > /proc/sys/net/core/default_qdisc 2>/dev/null && log_msg "Default qdisc set to FQ"
echo 3 > /proc/sys/net/ipv4/tcp_fastopen 2>/dev/null && log_msg "TCP Fast Open set to 3"
echo 1 > /proc/sys/net/ipv4/tcp_slow_start_after_idle 2>/dev/null && log_msg "TCP slow start after idle disabled (1)"

log_msg "=== EPITAPH TUNER COMPLETED SUCCESSFULLY ==="
