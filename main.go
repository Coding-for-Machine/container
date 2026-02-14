package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		buildContainer()
	case "shell":
		os.Args = append(os.Args[:2], "/bin/bash")
		run()
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Foydalanish: run <cmd> [args]")
			os.Exit(1)
		}
		run()
	case "child":
		child()
	default:
		fmt.Println("Noma'lum buyruq:", os.Args[1])
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║           MINI CONTAINER - 30 TIL VERSIYASI               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Foydalanish:")
	fmt.Println("  build               - Container rootfs yaratish (30 til)")
	fmt.Println("  shell               - Container ichida shell ochish")
	fmt.Println("  run <cmd> [args]    - Buyruqni container ichida bajarish")
	fmt.Println()
	fmt.Println("Misollar:")
	fmt.Println("  sudo ./minicontainer build")
	fmt.Println("  sudo ./minicontainer shell")
	fmt.Println("  sudo ./minicontainer run python3 --version")
}

func buildContainer() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║        🚀 MINI CONTAINER QURILISHI BOSHLANDI              ║")
	fmt.Println("║           30 ta til - Production Ready                    ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	os.MkdirAll("rootfs", 0755)

	fmt.Println("📦 Ubuntu base yuklanmoqda...")
	downloadUbuntuBase()

	fmt.Println()
	fmt.Println("🔧 30 ta til o'rnatilmoqda...")
	fmt.Println("   ⏱️  Bu 20-30 daqiqa davom etadi")
	fmt.Println("   ☕ Sabr qiling!")
	fmt.Println()

	installLanguagesInRootfs()

	fmt.Println()
	fmt.Println("🧹 Hajmni optimallashtirish...")
	optimizeRootfs()

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ CONTAINER TAYYOR!                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Ishlatish:")
	fmt.Println("   sudo ./minicontainer shell")
	fmt.Println()
	fmt.Println("Hajm:")
	cmd := exec.Command("du", "-sh", "rootfs")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func downloadUbuntuBase() {
	tarball := "/tmp/ubuntu-base.tar.gz"

	// Fayl validatsiyasi
	if stat, err := os.Stat(tarball); err == nil {
		fmt.Println("   ✓ Mavjud fayl tekshirilmoqda...")
		if stat.Size() < 20*1024*1024 {
			fmt.Println("   ⚠️  Fayl kichik, o'chirilmoqda...")
			os.Remove(tarball)
		} else {
			cmd := exec.Command("gzip", "-t", tarball)
			if err := cmd.Run(); err != nil {
				fmt.Println("   ⚠️  Fayl buzilgan, o'chirilmoqda...")
				os.Remove(tarball)
			} else {
				fmt.Println("   ✅ Fayl to'g'ri")
			}
		}
	}

	// Yuklab olish
	if _, err := os.Stat(tarball); os.IsNotExist(err) {
		fmt.Println("   📥 Ubuntu 22.04 yuklanmoqda...")
		for i := 1; i <= 3; i++ {
			fmt.Printf("   🔄 Urinish %d/3...\n", i)
			cmd := exec.Command("curl", "-L", "-C", "-", "--retry", "3",
				"--retry-delay", "2", "--max-time", "300", "-o", tarball,
				"http://cdimage.ubuntu.com/ubuntu-base/releases/22.04/release/ubuntu-base-22.04.5-base-amd64.tar.gz")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				if stat, err := os.Stat(tarball); err == nil && stat.Size() > 20*1024*1024 {
					fmt.Println("   ✅ Yuklash muvaffaqiyatli")
					break
				}
			}
			if i < 3 {
				fmt.Println("   ⚠️  Qayta urinish...")
				os.Remove(tarball)
				time.Sleep(2 * time.Second)
			} else {
				panic("Ubuntu base yuklanmadi")
			}
		}
	}

	// Rootfs validatsiyasi
	needsExtract := false
	if isEmpty("rootfs") {
		fmt.Println("   📁 rootfs bo'sh")
		needsExtract = true
	} else {
		fmt.Println("   🔍 rootfs tekshirilmoqda...")
		requiredFiles := []string{"rootfs/bin/bash", "rootfs/usr/bin", "rootfs/lib", "rootfs/etc"}
		isValid := true
		for _, file := range requiredFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				fmt.Printf("   ⚠️  %s topilmadi\n", file)
				isValid = false
				break
			}
		}
		if !isValid {
			fmt.Println("   ⚠️  rootfs noto'liq, qayta extract...")
			needsExtract = true
			exec.Command("sudo", "rm", "-rf", "rootfs").Run()
			time.Sleep(1 * time.Second)
			os.MkdirAll("rootfs", 0755)
		} else {
			fmt.Println("   ✅ rootfs to'g'ri")
		}
	}

	// Extract
	if needsExtract {
		fmt.Println("   📂 Extract qilinmoqda...")
		cmd := exec.Command("sudo", "tar", "-xzf", tarball, "-C", "rootfs")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("   ❌ Extract xatosi!")
			os.Remove(tarball)
			exec.Command("sudo", "rm", "-rf", "rootfs").Run()
			panic("Extract xatosi")
		}
		fmt.Println("   🔍 Validatsiya...")
		if _, err := os.Stat("rootfs/bin/bash"); os.IsNotExist(err) {
			panic("❌ /bin/bash topilmadi!")
		}
		fmt.Println("   ✅ Ubuntu base tayyor")
	}
}

func installLanguagesInRootfs() {
	mountForInstall()
	defer unmountForInstall()

	os.MkdirAll("rootfs/etc", 0755)
	os.WriteFile("rootfs/etc/resolv.conf", []byte("nameserver 8.8.8.8\nnameserver 1.1.1.1\n"), 0644)

	script := `#!/bin/bash
set -e
export DEBIAN_FRONTEND=noninteractive

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║    🚀 30 TA TIL - PROFESSIONAL PRODUCTION VERSION         ║"
echo "╚═══════════════════════════════════════════════════════════╝"

apt-get update -qq
apt-get upgrade -y -qq

# Asosiy paketlar
apt-get install -y -qq --no-install-recommends \
    ca-certificates curl wget git build-essential pkg-config \
    gnupg lsb-release software-properties-common unzip xz-utils

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⭐ CORE LANGUAGES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1-3. C/C++/Clang
echo "🔹 C/C++/Clang"
apt-get install -y -qq --no-install-recommends gcc g++ clang make
echo "   ✅ GCC: $(gcc --version | head -n1)"

# 4. Java
echo "🔹 Java"
apt-get install -y -qq --no-install-recommends openjdk-17-jdk
echo "   ✅ Java: $(java -version 2>&1 | head -n1)"

# 5. C#
echo "🔹 C#"
wget -q https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb
dpkg -i packages-microsoft-prod.deb
apt-get update -qq
apt-get install -y -qq --no-install-recommends dotnet-sdk-7.0
rm -f packages-microsoft-prod.deb
echo "   ✅ C#: $(dotnet --version)"

# 6-7. Python
echo "🔹 Python"
apt-get install -y -qq --no-install-recommends python3 python3-pip python3-dev pypy3
python3 -m pip install --break-system-packages --no-cache-dir numpy scipy
ln -sf /usr/bin/python3 /usr/bin/python
echo "   ✅ Python: $(python3 --version)"

# 8. Assembly
echo "🔹 Assembly"
apt-get install -y -qq --no-install-recommends nasm
echo "   ✅ NASM: $(nasm --version)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 WEB DEVELOPMENT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 9-10. Node.js + TypeScript
echo "🔹 Node.js + TypeScript"
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq nodejs
npm install -g npm@latest typescript ts-node
echo "   ✅ Node: $(node --version)"
echo "   ✅ TypeScript: $(tsc --version)"

# 11. PHP
echo "🔹 PHP"
apt-get install -y -qq --no-install-recommends php php-cli php-mbstring php-xml
echo "   ✅ PHP: $(php --version | head -n1)"

# 12. Ruby
echo "🔹 Ruby"
apt-get install -y -qq --no-install-recommends ruby ruby-dev
gem install bundler --no-document
echo "   ✅ Ruby: $(ruby --version)"

# 13. Perl
echo "🔹 Perl"
apt-get install -y -qq --no-install-recommends perl libperl-dev
echo "   ✅ Perl: $(perl --version | grep -o 'v[0-9.]*' | head -n1)"

# 14. Lua
echo "🔹 Lua"
apt-get install -y -qq --no-install-recommends lua5.4 luarocks
echo "   ✅ Lua: $(lua5.4 -v)"

# 15. Bash
echo "🔹 Bash"
echo "   ✅ Bash: $(bash --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 DATA SCIENCE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 16. R
echo "🔹 R"
apt-get install -y -qq --no-install-recommends r-base
echo "   ✅ R: $(R --version | head -n1)"

# 17. Julia
echo "🔹 Julia"
cd /tmp
wget -q https://julialang-s3.julialang.org/bin/linux/x64/1.9/julia-1.9.4-linux-x86_64.tar.gz
tar -xzf julia-1.9.4-linux-x86_64.tar.gz -C /opt
ln -sf /opt/julia-1.9.4/bin/julia /usr/local/bin/julia
rm -f julia-1.9.4-linux-x86_64.tar.gz
echo "   ✅ Julia: $(julia --version)"

# 18. Octave
echo "🔹 Octave"
apt-get install -y -qq --no-install-recommends octave
echo "   ✅ Octave: $(octave --version | head -n1)"

# 19. Fortran
echo "🔹 Fortran"
apt-get install -y -qq --no-install-recommends gfortran
echo "   ✅ Fortran: $(gfortran --version | head -n1)"

# 20. COBOL
echo "🔹 COBOL"
apt-get install -y -qq --no-install-recommends gnucobol
echo "   ✅ COBOL: $(cobc --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 SYSTEMS PROGRAMMING"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 21. Rust
echo "🔹 Rust"
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile=minimal
source $HOME/.cargo/env
cp -f $HOME/.cargo/bin/rustc /usr/local/bin/
cp -f $HOME/.cargo/bin/cargo /usr/local/bin/
echo "   ✅ Rust: $(rustc --version)"

# 22. Go
echo "🔹 Go"
cd /tmp
wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
rm -f go1.21.6.linux-amd64.tar.gz
ln -sf /usr/local/go/bin/go /usr/local/bin/go
ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
echo "   ✅ Go: $(go version)"

# 23. Zig
echo "🔹 Zig"
cd /tmp
wget -q https://ziglang.org/download/0.11.0/zig-linux-x86_64-0.11.0.tar.xz
tar -xf zig-linux-x86_64-0.11.0.tar.xz -C /opt
ln -sf /opt/zig-linux-x86_64-0.11.0/zig /usr/local/bin/zig
rm -f zig-linux-x86_64-0.11.0.tar.xz
echo "   ✅ Zig: $(zig version)"

# 24. D
echo "🔹 D"
cd /tmp
wget -q https://dlang.org/install.sh
bash install.sh dmd -p /opt/dlang || true
if [ -d /opt/dlang/dmd-* ]; then
    ln -sf /opt/dlang/dmd-*/linux/bin64/dmd /usr/local/bin/dmd
    echo "   ✅ D: $(dmd --version | head -n1)"
else
    echo "   ⚠️  D: O'rnatilmadi"
fi
rm -f install.sh

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎨 FUNCTIONAL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 25. Haskell
echo "🔹 Haskell"
apt-get install -y -qq --no-install-recommends haskell-platform
echo "   ✅ Haskell: $(ghc --version)"

# 26. Scala
echo "🔹 Scala"
apt-get install -y -qq --no-install-recommends scala
echo "   ✅ Scala: $(scala --version 2>&1 | head -n1)"

# 27. Elixir
echo "🔹 Elixir"
apt-get install -y -qq --no-install-recommends erlang elixir
echo "   ✅ Elixir: $(elixir --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 MOBILE & MODERN"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 28. Kotlin
echo "🔹 Kotlin"
cd /tmp
wget -q https://github.com/JetBrains/kotlin/releases/download/v1.9.20/kotlin-compiler-1.9.20.zip
unzip -q kotlin-compiler-1.9.20.zip -d /opt
ln -sf /opt/kotlinc/bin/kotlin /usr/local/bin/kotlin
ln -sf /opt/kotlinc/bin/kotlinc /usr/local/bin/kotlinc
rm -f kotlin-compiler-1.9.20.zip
echo "   ✅ Kotlin: $(kotlin -version 2>&1 | head -n1)"

# 29. Dart
echo "🔹 Dart"
wget -qO- https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /usr/share/keyrings/dart.gpg
echo 'deb [signed-by=/usr/share/keyrings/dart.gpg arch=amd64] https://storage.googleapis.com/download.dartlang.org/linux/debian stable main' > /etc/apt/sources.list.d/dart_stable.list
apt-get update -qq
apt-get install -y -qq --no-install-recommends dart
echo "   ✅ Dart: $(dart --version 2>&1 | head -n1)"

# 30. Swift
echo "🔹 Swift"
cd /tmp
wget -q https://download.swift.org/swift-5.9-release/ubuntu2204/swift-5.9-RELEASE/swift-5.9-RELEASE-ubuntu22.04.tar.gz
tar -xzf swift-5.9-RELEASE-ubuntu22.04.tar.gz -C /opt
ln -sf /opt/swift-5.9-RELEASE-ubuntu22.04/usr/bin/swift /usr/local/bin/swift
ln -sf /opt/swift-5.9-RELEASE-ubuntu22.04/usr/bin/swiftc /usr/local/bin/swiftc
rm -f swift-5.9-RELEASE-ubuntu22.04.tar.gz
echo "   ✅ Swift: $(swift --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "💾 DATABASE & TOOLS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

apt-get install -y -qq --no-install-recommends sqlite3 libsqlite3-dev
apt-get install -y -qq --no-install-recommends gdb valgrind make cmake
echo "✅ SQLite: $(sqlite3 --version)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧹 CLEANUP"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

apt-get clean
apt-get autoremove -y
rm -rf /var/lib/apt/lists/*
rm -rf /usr/share/doc /usr/share/man /usr/share/info
rm -rf /tmp/* /var/tmp/*
find /usr/lib -name "*.a" -delete 2>/dev/null || true
find /usr/local -name "*.pyc" -delete 2>/dev/null || true
rm -rf /root/.cache /root/.npm

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              ✅ 30 TA TIL TAYYOR!                         ║"
echo "╚═══════════════════════════════════════════════════════════╝"
`

	scriptPath := "rootfs/tmp/install.sh"
	must(os.WriteFile(scriptPath, []byte(script), 0755))

	fmt.Println("🔧 30 ta til o'rnatilmoqda...")
	cmd := exec.Command("sudo", "chroot", "rootfs", "/bin/bash", "/tmp/install.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ Xatolik: %v\n", err)
	}

	os.Remove(scriptPath)
}

func optimizeRootfs() {
	fmt.Println("   Optimallashtirish...")
	exec.Command("sudo", "rm", "-rf", "rootfs/usr/share/doc").Run()
	exec.Command("sudo", "rm", "-rf", "rootfs/usr/share/man").Run()
	exec.Command("sudo", "rm", "-rf", "rootfs/var/cache/apt").Run()
	exec.Command("sudo", "find", "rootfs/usr/lib", "-name", "*.a", "-delete").Run()
	fmt.Println("   ✅ Optimallash tugadi")
}

func run() {
	fmt.Printf("\n🌟 Host PID: %d\n", os.Getpid())
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())
}

func child() {
	fmt.Println("🐳 Container PID: 1")

	// Cgroup
	cg := "/sys/fs/cgroup/minicontainer"
	os.MkdirAll(cg, 0755)
	defer os.RemoveAll(cg)
	os.WriteFile(cg+"/memory.max", []byte("1000000000"), 0644)
	os.WriteFile(cg+"/cpu.max", []byte("100000 100000"), 0644)
	os.WriteFile(cg+"/cgroup.procs", []byte(fmt.Sprint(os.Getpid())), 0644)

	// Hostname
	must(syscall.Sethostname([]byte("minicontainer")))

	// ✅ MUHIM: Absolyut path olish chroot'dan OLDIN
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Working directory xatosi: %v\n", err)
		panic(err)
	}

	// Overlay setup
	fmt.Printf("📂 Working directory: %s\n", cwd)
	setupOverlayWithPath(cwd)

	// Chroot
	mergedPath := cwd + "/overlay/merged"
	if err := syscall.Chroot(mergedPath); err != nil {
		fmt.Printf("❌ Chroot xatosi: %v\n", err)
		panic(err)
	}
	must(os.Chdir("/"))

	// Limits
	setRlimits()

	// Mount proc, sys, tmp
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	must(syscall.Mount("sys", "/sys", "sysfs", 0, ""))
	must(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""))
	defer func() {
		syscall.Unmount("/proc", 0)
		syscall.Unmount("/sys", 0)
		syscall.Unmount("/tmp", 0)
	}()

	// Environment
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/root/.cargo/bin")
	os.Setenv("HOME", "/root")
	os.Setenv("TERM", "xterm-256color")

	// Run command
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
	}

	// Cleanup (chroot'dan keyin ishlamaydi, lekin defer'da bor)
	cleanupOverlay(cwd)
}

// ✅ TO'G'RILANGAN - Absolyut path bilan
func setupOverlayWithPath(cwd string) {
	base := cwd + "/rootfs"

	// Rootfs mavjudligini tekshirish
	if _, err := os.Stat(base); os.IsNotExist(err) {
		fmt.Printf("❌ Rootfs topilmadi: %s\n", base)
		panic(fmt.Errorf("rootfs not found at %s", base))
	}

	fmt.Printf("📂 Rootfs: %s\n", base)

	// Overlay kataloglarini yaratish
	lowerPath := cwd + "/overlay/lower"
	upperPath := cwd + "/overlay/upper"
	workPath := cwd + "/overlay/work"
	mergedPath := cwd + "/overlay/merged"

	os.MkdirAll(lowerPath, 0755)
	os.MkdirAll(upperPath, 0755)
	os.MkdirAll(workPath, 0755)
	os.MkdirAll(mergedPath, 0755)

	// Bind mount
	if err := syscall.Mount(base, lowerPath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		fmt.Printf("❌ Bind mount xatosi: %v\n", err)
		panic(err)
	}

	// Overlay mount
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerPath, upperPath, workPath)
	if err := syscall.Mount("overlay", mergedPath, "overlay", 0, opts); err != nil {
		fmt.Printf("❌ Overlay mount xatosi: %v\n", err)
		syscall.Unmount(lowerPath, 0)
		panic(err)
	}

	fmt.Println("✅ Overlay filesystem tayyor")
}

func cleanupOverlay(cwd string) {
	mergedPath := cwd + "/overlay/merged"
	lowerPath := cwd + "/overlay/lower"

	syscall.Unmount(mergedPath, syscall.MNT_DETACH)
	syscall.Unmount(lowerPath, syscall.MNT_DETACH)
	os.RemoveAll(cwd + "/overlay")
}

func setRlimits() {
	unix.Setrlimit(unix.RLIMIT_NPROC, &unix.Rlimit{Cur: 64, Max: 64})
	unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: 256, Max: 256})
	unix.Setrlimit(unix.RLIMIT_CPU, &unix.Rlimit{Cur: 2, Max: 2})
	unix.Setrlimit(unix.RLIMIT_STACK, &unix.Rlimit{Cur: 8 * 1024 * 1024, Max: 8 * 1024 * 1024})
	unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0})
}

func mountForInstall() {
	os.MkdirAll("rootfs/proc", 0755)
	os.MkdirAll("rootfs/sys", 0755)
	os.MkdirAll("rootfs/dev", 0755)
	os.MkdirAll("rootfs/tmp", 0755)

	exec.Command("sudo", "mount", "-t", "proc", "proc", "rootfs/proc").Run()
	exec.Command("sudo", "mount", "-t", "sysfs", "sys", "rootfs/sys").Run()
	exec.Command("sudo", "mount", "--bind", "/dev", "rootfs/dev").Run()
	exec.Command("sudo", "mount", "--bind", "/tmp", "rootfs/tmp").Run()
	if _, err := os.Stat("rootfs/etc/resolv.conf"); err == nil {
		exec.Command("sudo", "mount", "--bind", "/etc/resolv.conf", "rootfs/etc/resolv.conf").Run()
	}
}

func unmountForInstall() {
	exec.Command("sudo", "umount", "-l", "rootfs/etc/resolv.conf").Run()
	exec.Command("sudo", "umount", "-l", "rootfs/tmp").Run()
	exec.Command("sudo", "umount", "-l", "rootfs/dev").Run()
	exec.Command("sudo", "umount", "-l", "rootfs/sys").Run()
	exec.Command("sudo", "umount", "-l", "rootfs/proc").Run()
}

func isEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return true
	}
	return len(entries) == 0
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}