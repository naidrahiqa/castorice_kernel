import os
import re

def main():
    print("⏳ Running Schedutil Rate Limit Unlocker...")
    filepath = "kernel/sched/cpufreq_schedutil.c"
    if not os.path.exists(filepath):
        filepath = "common/kernel/sched/cpufreq_schedutil.c"
    if not os.path.exists(filepath):
        filepath = "../kernel/common/kernel/sched/cpufreq_schedutil.c"
    
    if os.path.exists(filepath):
        print(f"🔍 Found cpufreq_schedutil.c at: {filepath}")
        with open(filepath, "r") as f:
            content = f.read()

        # Unlock rate limit boundaries so userspace can set values down to 100 microseconds
        # replacing various typical checks with val < 100
        content = re.sub(r'val\s*<\s*TICK_USEC\s*/\s*2', 'val < 100', content)
        content = re.sub(r'val\s*<\s*TICK_USEC', 'val < 100', content)
        content = re.sub(r'val\s*<\s*1000', 'val < 100', content)
        content = re.sub(r'val\s*<\s*500', 'val < 100', content)
        content = re.sub(r'val\s*<\s*MIN_RATE_LIMIT_US', 'val < 100', content)

        with open(filepath, "w") as f:
            f.write(content)
        print("✅ Successfully patched cpufreq_schedutil.c to unlock 100µs limit!")
    else:
        print("⚠️ Warning: cpufreq_schedutil.c not found!")

if __name__ == "__main__":
    main()
