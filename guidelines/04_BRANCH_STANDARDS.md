# Guidelines: KernelSU-Next Branch & Versioning Standards

> [!IMPORTANT]
> This document defines the absolute, unalterable gold standards for KernelSU-Next repository sources and branches for the Epitaph kernel. **These standards must never be modified** by any future developer or AI agent.

---

## 📌 Branch Matrix & Sources

| Build Type | Git Repository | Branch | Purpose |
| :--- | :--- | :---: | :--- |
| **SUSFS Build** | `https://github.com/pershoot/KernelSU-Next` | **`dev-susfs`** | Pre-patched SUSFS v2.1.0 driver layers to ensure seamless safety and symbol mapping. |
| **No-SUSFS Build** | `https://github.com/KernelSU-Next/KernelSU-Next` | **`dev`** | Upstream GKI 6.6 compatible development branch. |

---

## ⚠️ Gold Standards of Stability (DO NOT VIOLATE)

1. **Never Change to `next`, `next-susfs`, or `main`**
   * **Rule**: The branches specified above (**`dev-susfs`** and **`dev`**) are verified to be the only branches compatible with GKI 6.6 compile rules for this kernel.
   * **Rationale**: Hulu upstream resmi `KernelSU-Next` tidak menyediakan branch bernama `next` atau `main` untuk rilis aktif GKI 6.6. Mengubah branch ke nama lain akan memicu kegagalan kloning fatal (`fatal: Remote branch not found`) pada runner CI/CD.

2. **Commit Staged Changes Before Bazel Compilation**
   * **Rule**: Selalu jalankan `git commit` setelah melakukan `git add` pada semua modifikasi yang dibuat oleh skrip persiapan.
   * **Rationale**: Bazel Kleaf sandbox hanya melacak berkas yang sudah masuk ke riwayat komit Git HEAD. Modifikasi mentah (staged/unstaged) tidak akan terlihat oleh sandbox, yang dapat memicu kegagalan pemetaan modul vendor Xiaomi.

3. **Ignore Manual `10_enable_susfs_for_ksu.patch` for `dev-susfs`**
   * **Rule**: Lewati penambalan manual KSU-side patch jika menggunakan fork pershoot.
   * **Rationale**: Fork pershoot branch `dev-susfs` sudah memiliki integrasi pra-penambalan driver SUSFS yang andal secara bawaan, sehingga tidak memerlukan patch mandiri dari modul pengguna.

---
*Created & Maintained by Epitaph Core Team & Antigravity AI (May 2026)*
