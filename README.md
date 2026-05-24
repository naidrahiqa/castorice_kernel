<div align="center">

<!--
  EPITAPH KERNEL — README.md
  Banner: SVG inline, renders natively on GitHub
-->

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 860 200" width="860" height="200">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#0a0a0f"/>
      <stop offset="50%" style="stop-color:#0d1117"/>
      <stop offset="100%" style="stop-color:#0a0a0f"/>
    </linearGradient>
    <linearGradient id="accent" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#6366f1"/>
      <stop offset="50%" style="stop-color:#8b5cf6"/>
      <stop offset="100%" style="stop-color:#06b6d4"/>
    </linearGradient>
    <linearGradient id="titlegrad" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#e2e8f0"/>
      <stop offset="60%" style="stop-color:#ffffff"/>
      <stop offset="100%" style="stop-color:#94a3b8"/>
    </linearGradient>
    <filter id="glow">
      <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
    <filter id="softglow">
      <feGaussianBlur stdDeviation="8" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>

  <!-- Background -->
  <rect width="860" height="200" fill="url(#bg)" rx="12"/>

  <!-- Subtle grid pattern -->
  <g opacity="0.04" stroke="#6366f1" stroke-width="0.5">
    <line x1="0" y1="40" x2="860" y2="40"/>
    <line x1="0" y1="80" x2="860" y2="80"/>
    <line x1="0" y1="120" x2="860" y2="120"/>
    <line x1="0" y1="160" x2="860" y2="160"/>
    <line x1="100" y1="0" x2="100" y2="200"/>
    <line x1="200" y1="0" x2="200" y2="200"/>
    <line x1="300" y1="0" x2="300" y2="200"/>
    <line x1="400" y1="0" x2="400" y2="200"/>
    <line x1="500" y1="0" x2="500" y2="200"/>
    <line x1="600" y1="0" x2="600" y2="200"/>
    <line x1="700" y1="0" x2="700" y2="200"/>
    <line x1="800" y1="0" x2="800" y2="200"/>
  </g>

  <!-- Decorative orb left -->
  <circle cx="60" cy="100" r="80" fill="#6366f1" opacity="0.06" filter="url(#softglow)"/>
  <!-- Decorative orb right -->
  <circle cx="800" cy="100" r="90" fill="#06b6d4" opacity="0.05" filter="url(#softglow)"/>

  <!-- Top accent line -->
  <rect x="0" y="0" width="860" height="3" fill="url(#accent)" rx="2" filter="url(#glow)"/>
  <!-- Bottom accent line -->
  <rect x="0" y="197" width="860" height="3" fill="url(#accent)" rx="2" opacity="0.5"/>

  <!-- Left vertical accent -->
  <rect x="50" y="40" width="2" height="120" fill="url(#accent)" opacity="0.6" rx="1"/>

  <!-- Kernel icon / monogram -->
  <text x="72" y="115" font-family="monospace" font-size="42" fill="url(#accent)" opacity="0.9" filter="url(#glow)" font-weight="bold">∂</text>

  <!-- Main title -->
  <text x="130" y="95" font-family="Georgia, 'Times New Roman', serif" font-size="52" font-weight="bold" fill="url(#titlegrad)" letter-spacing="3" filter="url(#glow)">EPITAPH</text>
  <text x="131" y="126" font-family="'Courier New', monospace" font-size="14" fill="#64748b" letter-spacing="8">K  E  R  N  E  L</text>

  <!-- Divider -->
  <rect x="130" y="138" width="280" height="1" fill="url(#accent)" opacity="0.4"/>

  <!-- Subtitle -->
  <text x="130" y="158" font-family="'Courier New', monospace" font-size="11" fill="#475569" letter-spacing="2">GKI 6.6  ·  ANDROID 15  ·  REDMI 12 (fire)</text>

  <!-- Right side: specs -->
  <g transform="translate(560, 50)">
    <rect width="250" height="110" rx="8" fill="#ffffff" opacity="0.03" stroke="#334155" stroke-width="0.5"/>
    <text x="16" y="24" font-family="'Courier New', monospace" font-size="10" fill="#6366f1" letter-spacing="1">// SPECS</text>
    <text x="16" y="44" font-family="'Courier New', monospace" font-size="10" fill="#94a3b8">ROOT    <tspan fill="#e2e8f0">KernelSU-Next</tspan></text>
    <text x="16" y="60" font-family="'Courier New', monospace" font-size="10" fill="#94a3b8">GOV     <tspan fill="#e2e8f0">schedutil</tspan></text>
    <text x="16" y="76" font-family="'Courier New', monospace" font-size="10" fill="#94a3b8">TCP     <tspan fill="#e2e8f0">BBR + FQ</tspan></text>
    <text x="16" y="92" font-family="'Courier New', monospace" font-size="10" fill="#94a3b8">ZRAM    <tspan fill="#e2e8f0">ZSTD multi-stream</tspan></text>
  </g>
</svg>

<br/>

<!-- PRIMARY BADGES -->
![Android](https://img.shields.io/badge/Android-15-3DDC84?style=for-the-badge&logo=android&logoColor=white)
![Device](https://img.shields.io/badge/Redmi_12-fire-FF6900?style=for-the-badge&logo=xiaomi&logoColor=white)
![Chipset](https://img.shields.io/badge/Helio_G88-MT6769-orange?style=for-the-badge)
![Kernel](https://img.shields.io/badge/GKI-6.6-0ea5e9?style=for-the-badge&logo=linux&logoColor=white)

<br/>

<!-- STATUS BADGES -->
![Build](https://img.shields.io/github/actions/workflow/status/naidrahiqa/epitaph_kernel/build_manager_gki.yml?style=flat-square&label=CI%2FCD&logo=githubactions&logoColor=white)
![Release](https://img.shields.io/github/v/release/naidrahiqa/epitaph_kernel?style=flat-square&label=Latest&color=8b5cf6)
![Stars](https://img.shields.io/github/stars/naidrahiqa/epitaph_kernel?style=flat-square&color=fbbf24&logo=github)
![License](https://img.shields.io/badge/License-GPL--2.0-22c55e?style=flat-square)
![KernelSU](https://img.shields.io/badge/KernelSU--Next-supported-6366f1?style=flat-square)
![SUSFS](https://img.shields.io/badge/SUSFS-optional-64748b?style=flat-square)

</div>

---

## 📖 Overview

**Epitaph** is a custom GKI 6.6 kernel for **Redmi 12 (fire)** running **Android 15 HyperOS 2.0**, built from Google's `common-android15-6.6` branch with a focus on stability, root compatibility, and everyday performance.

Compiled automatically via a multi-toolchain CI/CD pipeline on GitHub Actions, Epitaph ships with kernel-level root integration (KernelSU-Next), optional root-hiding (SUSFS), WiFi/Hotspot fixes, and a post-boot tuner that optimizes the device without touching stock vendor modules.

> **If you're here to flash:** jump straight to [Installation](#-installation).
> **If you're a developer:** check [Architecture](#-architecture--ci-cd) and [Contributing](#-contributing).

---

## ✨ Features

### 🔐 Root & Security
| Feature | Status | Notes |
|---|---|---|
| KernelSU-Next | ✅ Built-in | Kernel-level, no Magisk needed |
| SUSFS 4 KSU | ✅ Optional build | Hides root from banking apps |
| Vermagic bypass | ✅ Always on | Stock Xiaomi modules load cleanly |

### ⚡ Performance
| Feature | Status | Notes |
|---|---|---|
| CPU Governor | ✅ schedutil | Tuned rate limits via post-boot script |
| TCP Congestion | ✅ BBR | Better ping, more stable throughput |
| Packet Scheduler | ✅ FQ | Smoother network under load |
| I/O Scheduler | ✅ BFQ + Kyber | App launches faster under background I/O |
| Timer Frequency | ✅ HZ=300 | Balanced responsiveness & battery |
| WireGuard | ✅ Built-in | Zero overhead VPN at kernel level |

### 🧠 Memory
| Feature | Status | Notes |
|---|---|---|
| ZRAM ZSTD | ✅ Active | ~25% better compression vs LZ4 |
| ZRAM multi-stream | ✅ Active | Parallel compression on all cores |
| MGLRU | ✅ Active | Smarter background app retention |
| Swappiness tuning | ✅ Post-boot | 180 via Epitaph Tuner |

### 📶 Connectivity
| Feature | Status | Notes |
|---|---|---|
| WiFi modules | ✅ cfg80211 + mac80211 | Modular, AnyKernel3 shipped |
| Hotspot (IPv4) | ✅ Full NAT | Netfilter Masquerade |
| Hotspot (IPv6) | ✅ Full NAT | ip6tables Masquerade |
| WiFi fallback loader | ✅ Post-boot | Auto-insmod if systemless fails |

### 🔧 Debugging & Recovery
| Feature | Status | Notes |
|---|---|---|
| PStore / RAMoops | ✅ Active | Crash logs survive reboot |
| Rescue boot.img | ✅ Separate build | Always-boots fallback for log pull |
| Telegram error dump | ✅ Automatic | Build errors auto-sent to error channel |

---

## 📱 Device Compatibility

| Field | Value |
|---|---|
| **Device** | Xiaomi Redmi 12 |
| **Codename** | `fire` |
| **Chipset** | MediaTek Helio G88 (MT6769) |
| **ROM** | HyperOS 2.0 (Android 15) |
| **Architecture** | ARM64 / GKI 6.6 |
| **KMI** | android15-8 |

> ⚠️ **This kernel will not work on any other device or ROM version.** AnyKernel3 enforces `supported.versions=15` and `device.name=fire`. Installation will abort automatically on unsupported configurations.

---

## 📥 Installation

### Prerequisites

- Device on **HyperOS 2.0 (Android 15)** — no other ROM supported
- [KernelFlasher](https://github.com/capntrips/KernelFlasher/releases) installed with root access
- A backup of your stock `boot.img` (extract from your HyperOS Fastboot ROM)
- Fastboot ROM matching your current firmware version (for vbmeta files)

### Step 1 — Disable AVB (One-time only)

Boot into Fastboot mode (`Volume Down + Power`) and run:

```bash
fastboot --disable-verity --disable-verification flash vbmeta vbmeta.img
fastboot --disable-verity --disable-verification flash vbmeta_system vbmeta_system.img
fastboot --disable-verity --disable-verification flash vbmeta_vendor vbmeta_vendor.img
```

> Extract `vbmeta.img` files from the official HyperOS 2.0 Fastboot ROM matching your firmware.

### Step 2 — Choose your build

<div align="center">

| Build Variant | Root | SUSFS | Best for |
|---|---|---|---|
| `Epitaph-...-kernelsu-next-false` | KernelSU-Next | ❌ | Daily driver, gaming |
| `Epitaph-...-kernelsu-next-SUSFS` | KernelSU-Next | ✅ | Banking apps, root hiding |

</div>

### Step 3 — Flash

1. Download the AnyKernel3 ZIP from [**Latest Release →**](../../releases/latest)
2. Open **KernelFlasher**, select the ZIP, tap **Flash**
3. Reboot — done

> First boot may take ~10 seconds longer than usual as the post-boot tuner runs.

---

## 🚨 Recovery (If Bootloop)

**Don't panic.** Follow this exact sequence:

```bash
# Step 1: Flash stock boot via Fastboot
fastboot flash boot boot_stock.img
fastboot reboot
```

Once back on stock ROM with WiFi working:

```bash
# Step 2: Pull crash log from the failed kernel (from PC terminal)
adb shell "su -c cat /sys/fs/pstore/console-ramoops-0" > last_kmsg.txt
```

Then re-flash a stable Epitaph ZIP via KernelFlasher once the issue is identified.

> **Never** flash multiple boot images via Fastboot back-to-back — this wipes the RAMoops crash log.

---

## 🏗️ Architecture & CI/CD

Epitaph uses a multi-stage GitHub Actions pipeline:

```
build_manager_gki.yml          ← Dispatcher (matrix: toolchain × SUSFS variant)
        │
        ├── _build_kernel_core.yml    ← Core build (per matrix entry)
        │       ├── prepare_kernel_build.sh   (disk, deps, repo sync, KSU setup)
        │       ├── Setup SUSFS (optional)
        │       ├── Configure Kernel (defconfig)
        │       ├── Build (Bazel or custom toolchain)
        │       ├── Package AnyKernel3
        │       └── Release + Telegram notify
        │
        └── build_debug_bootimg.yml   ← Rescue kernel (always-boots, PStore enabled)
```

### Supported Toolchains

| Toolchain | Type | Notes |
|---|---|---|
| `bazel-default` | Bazel/Kleaf | **Recommended.** AOSP official build system |
| `aosp-latest` | make | AOSP Clang from crdroidandroid |
| `zyc-latest` | make | ZyClang, experimental |
| `weebx-latest` | make | WeebX Clang, experimental |
| `neutron-latest` | make | Neutron Clang, experimental |

> ⚠️ Only `bazel-default` is production-tested. Custom toolchains may produce bootloops and are for experimentation only.

---

## 🎛️ Epitaph Tuner

A post-boot optimization script ships inside AnyKernel3 and installs to `/data/adb/service.d/epitaph_tuner.sh`. It runs automatically at every boot via KernelSU/Magisk service.d.

**What it does:**

```
1. WiFi fallback loader    — insmod cfg80211/mac80211 if systemless loading fails
2. CPU schedutil tuning    — up_rate=500µs, down_rate=10000µs for smooth UI
3. GPU limiter reset       — bypasses MTK GED thermal throttle bug
4. VM tuning               — swappiness=180, dirty_ratio=20, vfs_cache_pressure=100
5. ZRAM streams            — max_comp_streams=2 for parallel compression
6. Read-ahead boost        — 512KB for faster app load times
7. TCP tuning              — BBR congestion, FQ qdisc, tcp_fastopen=3
```

Check tuner status anytime:

```bash
adb shell "cat /data/local/tmp/epitaph_tuner.log"
```

---

## 📊 Build Variants Matrix

When triggered with `susfs_variant=both` and `toolchain=all`, the CI produces up to 10 builds in parallel:

<div align="center">

| | Bazel | AOSP | ZyClang | WeebX | Neutron |
|---|---|---|---|---|---|
| **No SUSFS** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **+ SUSFS** | ✅ | ✅ | ✅ | ✅ | ✅ |

</div>

Each build produces a standalone AnyKernel3 ZIP named:
```
Epitaph-{Toolchain}-kernelsu-next[-SUSFS]-{Date}-AnyKernel3.zip
```

---

## 🛠️ Developer Notes

### Building locally (via GitHub Actions)

1. Fork this repository
2. Go to **Actions → 🎛️ GKI Control Center**
3. Click **Run workflow**, choose your variant
4. Download the artifact from the completed run

### Adding patches

Drop `.patch` files into the `patches/` directory. They are applied automatically during `prepare_kernel_build.sh` via `patch -p1`.

### Key scripts

| Script | Purpose |
|---|---|
| `scripts/prepare_kernel_build.sh` | Disk cleanup, deps, repo sync, KSU integration |
| `scripts/epitaph_tuner.sh` | Post-boot runtime optimizations |
| `workflow_scripts/patch_build_system.py` | Registers WiFi modules in Bazel BUILD.bazel |
| `workflow_scripts/patch_vermagic.py` | Bypasses vermagic mismatch for stock vendor modules |
| `workflow_scripts/patch_kbuild.py` | Injects static KSU version into Kbuild |

---

## 👥 Team

| Contributor | Role |
|---|---|
| **Faqih Ardian Syah ([@naidrahiqa](https://github.com/naidrahiqa))** | Lead Maintainer & Kernel Architect |
| **Antigravity AI** | CI/CD Engineer & Autonomous Co-Developer |
| **Gemini** (Google) | Architectural Reasoning |
| **Claude** (Anthropic) | Code Review & Syntax Refinement |
| **DeepSeek** | Performance Optimization |
| **Qwen** (Alibaba) | Bazel Build System |

---

## 📄 License

This project is licensed under **GPL-2.0** in accordance with the Linux kernel license.
Vendor blobs and proprietary modules remain property of their respective owners.

---

<div align="center">

*Built with precision. Flashed with confidence.*

**[⬇️ Download Latest Release](../../releases/latest)** &nbsp;·&nbsp; **[🐛 Report Issue](../../issues/new)** &nbsp;·&nbsp; **[📋 Changelog](../../releases)**

<br/>

![Footer](https://img.shields.io/badge/Epitaph_Kernel-Redmi_12-6366f1?style=for-the-badge&logo=linux&logoColor=white)

</div>