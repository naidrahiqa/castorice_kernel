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

# 1. FALLBACK WIFI MODULE LOADER
# Ensures 100% WiFi/Hotspot reliability if systemless module loading fails
if ! lsmod | grep -q cfg80211; then
  if [ -d /data/adb/wifi_fix ]; then
    insmod /data/adb/wifi_fix/rfkill.ko 2>/dev/null
    insmod /data/adb/wifi_fix/libarc4.ko 2>/dev/null
    insmod /data/adb/wifi_fix/cfg80211.ko 2>/dev/null
    insmod /data/adb/wifi_fix/mac80211.ko 2>/dev/null
  fi
fi

# 2. CPU SCHEDUTIL GOVERNOR OPTIMIZATIONS
# Smooths UI transitions & eliminates micro-stutters
for policy in /sys/devices/system/cpu/cpufreq/policy*; do
  if [ -f "$policy/scaling_governor" ]; then
    gov=$(cat "$policy/scaling_governor")
    if [ "$gov" = "schedutil" ]; then
      echo 500 > "$policy/schedutil/up_rate_limit_us" 2>/dev/null
      echo 10000 > "$policy/schedutil/down_rate_limit_us" 2>/dev/null
    fi
  fi
done

# 3. GPU GED & MALI LIMITER RESET
# Fixes GPU frequency locks / MTK thermal driver throttle bugs
if [ -d /sys/kernel/ged/hal ]; then
  echo 1 > /sys/kernel/ged/hal/gpu_boost 2>/dev/null
  echo 1 > /sys/module/ged/parameters/boost_gpu_enable 2>/dev/null
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
  fi
done

# 4. MEMORY & VIRTUAL MEMORY TUNING
# Resolves high RAM consumption, prevents OOM & background app kills
echo 180 > /proc/sys/vm/swappiness 2>/dev/null
echo 100 > /proc/sys/vm/vfs_cache_pressure 2>/dev/null
echo 20 > /proc/sys/vm/dirty_ratio 2>/dev/null
echo 5 > /proc/sys/vm/dirty_background_ratio 2>/dev/null
if [ -e /dev/block/zram0 ]; then
  echo 2 > /sys/block/zram0/max_comp_streams 2>/dev/null
fi

# 5. STORAGE READ-AHEAD OPTIMIZATION
# Enhances app loading speed
for queue in /sys/block/*/queue; do
  if [ -d "$queue" ]; then
    echo 512 > "$queue/read_ahead_kb" 2>/dev/null
  fi
done
