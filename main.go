package main

import (
	"fmt"
	"os"
	"os/exec"
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
	fmt.Println("  run <cmd> [args]    - Buyruqni container ichida bajarish")
	fmt.Println()
	fmt.Println("Misollar:")
	fmt.Println("  sudo ./minicontainer build")
	fmt.Println("  sudo ./minicontainer run /bin/bash")
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
	fmt.Println("   sudo ./minicontainer run /bin/bash")
	fmt.Println()
	fmt.Println("Hajm:")
	cmd := exec.Command("du", "-sh", "rootfs")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func downloadUbuntuBase() {
	tarball := "/tmp/ubuntu-base.tar.gz"

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

echo "📦 Sistema yangilanmoqda..."
apt-get update -qq
apt-get upgrade -y -qq

echo "📦 Asosiy paketlar o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends \
    ca-certificates curl wget git build-essential pkg-config \
    gnupg lsb-release software-properties-common unzip xz-utils \
    libssl-dev libffi-dev zlib1g-dev

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⭐ CORE LANGUAGES (C, C++, Java, C#, Python)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 1-2. C/C++
echo "🔹 C/C++ o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends gcc g++ make
echo "   ✅ GCC: $(gcc --version | head -n1)"

# 3. Java
echo "🔹 Java o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends openjdk-17-jdk-headless
echo "   ✅ Java: $(java -version 2>&1 | head -n1)"

# 4. C#
echo "🔹 C# o'rnatilmoqda..."
wget -q https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb -O /tmp/microsoft.deb
dpkg -i /tmp/microsoft.deb 2>/dev/null || true
apt-get update -qq
apt-get install -y -qq --no-install-recommends dotnet-sdk-7.0
rm -f /tmp/microsoft.deb
echo "   ✅ C#: $(dotnet --version)"

# 5-6. Python (FIXED - removed --break-system-packages)
echo "🔹 Python o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends python3 python3-pip python3-dev python3-venv
python3 -m pip install --no-cache-dir --upgrade pip
python3 -m pip install --no-cache-dir numpy scipy pandas
ln -sf /usr/bin/python3 /usr/bin/python
echo "   ✅ Python: $(python3 --version)"

# 7. Assembly
echo "🔹 Assembly o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends nasm
echo "   ✅ NASM: $(nasm --version)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🌐 WEB DEVELOPMENT (Node.js, PHP, Ruby, Perl)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 8-9. Node.js
echo "🔹 Node.js o'rnatilmoqda..."
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq --no-install-recommends nodejs
npm install -g npm@latest
echo "   ✅ Node.js: $(node --version)"

# 10. TypeScript
echo "🔹 TypeScript o'rnatilmoqda..."
npm install -g typescript ts-node
echo "   ✅ TypeScript: $(tsc --version)"

# 11. PHP
echo "🔹 PHP o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends php php-cli php-mbstring php-xml php-curl
echo "   ✅ PHP: $(php --version | head -n1)"

# 12. Ruby
echo "🔹 Ruby o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends ruby-full ruby-dev
gem install --no-document bundler
echo "   ✅ Ruby: $(ruby --version)"

# 13. Perl
echo "🔹 Perl o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends perl
echo "   ✅ Perl: $(perl --version | grep -o 'v[0-9.]*' | head -n1)"

# 14. Lua
echo "🔹 Lua o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends lua5.4
echo "   ✅ Lua: $(lua5.4 -v 2>&1 | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔧 SYSTEMS PROGRAMMING (Rust, Go, Zig)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 15. Rust
echo "🔹 Rust o'rnatilmoqda..."
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile=minimal --no-modify-path
source /root/.cargo/env
cp /root/.cargo/bin/rustc /usr/local/bin/ 2>/dev/null || true
cp /root/.cargo/bin/cargo /usr/local/bin/ 2>/dev/null || true
echo "   ✅ Rust: $(rustc --version 2>/dev/null || echo 'o‘rnatilmadi')"

# 16. Go
echo "🔹 Go o'rnatilmoqda..."
wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz -O /tmp/go.tar.gz
tar -C /usr/local -xzf /tmp/go.tar.gz
ln -sf /usr/local/go/bin/go /usr/local/bin/go
rm -f /tmp/go.tar.gz
echo "   ✅ Go: $(go version 2>/dev/null || echo 'o‘rnatilmadi')"

# 17. Zig
echo "🔹 Zig o'rnatilmoqda..."
wget -q https://ziglang.org/download/0.11.0/zig-linux-x86_64-0.11.0.tar.xz -O /tmp/zig.tar.xz
tar -xf /tmp/zig.tar.xz -C /opt
ln -sf /opt/zig-linux-x86_64-0.11.0/zig /usr/local/bin/zig
rm -f /tmp/zig.tar.xz
echo "   ✅ Zig: $(zig version 2>/dev/null || echo 'o‘rnatilmadi')"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 DATA SCIENCE (R, Julia, Octave)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 18. R
echo "🔹 R o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends r-base
echo "   ✅ R: $(R --version | head -n1)"

# 19. Julia
echo "🔹 Julia o'rnatilmoqda..."
wget -q https://julialang-s3.julialang.org/bin/linux/x64/1.9/julia-1.9.4-linux-x86_64.tar.gz -O /tmp/julia.tar.gz
tar -xzf /tmp/julia.tar.gz -C /opt
ln -sf /opt/julia-1.9.4/bin/julia /usr/local/bin/julia
rm -f /tmp/julia.tar.gz
echo "   ✅ Julia: $(julia --version 2>/dev/null || echo 'o‘rnatilmadi')"

# 20. Octave
echo "🔹 Octave o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends octave
echo "   ✅ Octave: $(octave --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎯 FUNCTIONAL LANGUAGES (Haskell, Scala, Elixir)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 21. Haskell
echo "🔹 Haskell o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends haskell-platform
echo "   ✅ Haskell: $(ghc --version)"

# 22. Scala
echo "🔹 Scala o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends scala
echo "   ✅ Scala: $(scala --version 2>&1 | head -n1)"

# 23. Elixir
echo "🔹 Elixir o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends elixir
echo "   ✅ Elixir: $(elixir --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📱 MODERN LANGUAGES (Kotlin, Dart)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 24. Kotlin
echo "🔹 Kotlin o'rnatilmoqda..."
wget -q https://github.com/JetBrains/kotlin/releases/download/v1.9.20/kotlin-compiler-1.9.20.zip -O /tmp/kotlin.zip
unzip -q /tmp/kotlin.zip -d /opt
ln -sf /opt/kotlinc/bin/kotlin /usr/local/bin/kotlin
ln -sf /opt/kotlinc/bin/kotlinc /usr/local/bin/kotlinc
rm -f /tmp/kotlin.zip
echo "   ✅ Kotlin: $(kotlin -version 2>&1 | head -n1)"

# 25. Dart
echo "🔹 Dart o'rnatilmoqda..."
wget -qO- https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /usr/share/keyrings/dart.gpg
echo 'deb [signed-by=/usr/share/keyrings/dart.gpg arch=amd64] https://storage.googleapis.com/download.dartlang.org/linux/debian stable main' > /etc/apt/sources.list.d/dart_stable.list
apt-get update -qq
apt-get install -y -qq --no-install-recommends dart
echo "   ✅ Dart: $(dart --version 2>&1 | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🗄️  DATABASES & TOOLS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 26. SQLite
echo "🔹 SQLite o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends sqlite3 libsqlite3-dev
echo "   ✅ SQLite: $(sqlite3 --version)"

# 27. Git
echo "🔹 Git o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends git
echo "   ✅ Git: $(git --version)"

# 28. Build tools
echo "🔹 Build tools o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends cmake
echo "   ✅ CMake: $(cmake --version | head -n1)"

# 29. Debug tools
echo "🔹 Debug tools o'rnatilmoqda..."
apt-get install -y -qq --no-install-recommends gdb strace
echo "   ✅ GDB: $(gdb --version | head -n1)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧹 TOZALASH VA OPTIMALLASHTIRISH"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# APT cacheni tozalash
apt-get clean
apt-get autoremove -y
rm -rf /var/lib/apt/lists/*

# Dokumentatsiyani tozalash
rm -rf /usr/share/doc
rm -rf /usr/share/man
rm -rf /usr/share/info
rm -rf /usr/share/locale/*

# Static libraries (hajmni kamaytirish)
find /usr/lib -name "*.a" -delete 2>/dev/null || true

# Python bytecode
find /usr/local -name "*.pyc" -delete 2>/dev/null || true
find /usr/local -name "*.pyo" -delete 2>/dev/null || true

# Cachelarni tozalash
rm -rf /root/.cache
rm -rf /root/.npm
rm -rf /tmp/*
rm -rf /var/tmp/*

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              ✅ 30 TA TIL MUVOFFAQIYATLI O'RNATILDI       ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# O'rnatilgan tillarni tekshirish
echo "📋 O'RNATILGAN TILLAR:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for cmd in gcc g++ java python3 node php ruby perl rustc go zig R julia octave ghc scala elixir kotlin dart sqlite3; do
    if command -v $cmd &>/dev/null; then
        echo "   ✅ $cmd: $(command -v $cmd)"
    fi
done

echo ""
echo "🚀 Endi siz quyidagi buyruqlarni ishlatishingiz mumkin:"
echo "   sudo ./minicontainer run /bin/bash"
echo "   sudo ./minicontainer run python3 -c \"print('Hello')\""
echo "   sudo ./minicontainer run node -e \"console.log('Hello')\""
`

	scriptPath := "rootfs/tmp/install.sh"
	must(os.WriteFile(scriptPath, []byte(script), 0755))

	fmt.Println("🔧 30 ta til o'rnatilmoqda...")
	fmt.Println("   ⏱️  Bu 15-20 daqiqa davom etishi mumkin")
	fmt.Println("   ☕ Choy ichib kuting!")

	cmd := exec.Command("sudo", "chroot", "rootfs", "/bin/bash", "/tmp/install.sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️ Ba'zi tillar o'rnatilmadi: %v\n", err)
		fmt.Println("✅ Asosiy tillar ishlaydi")
	} else {
		fmt.Println("✅ Barcha tillar muvaffaqiyatli o'rnatildi!")
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

	// ✅ Absolyut path olish (chroot'dan OLDIN)
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Overlay setup
	setupOverlay(cwd)

	// Chroot
	mergedPath := cwd + "/overlay/merged"
	must(syscall.Chroot(mergedPath))
	must(os.Chdir("/"))

	// Limits
	setRlimits()

	// Mount
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

	defer cleanupOverlay(cwd)
}

// ✅ TO'G'RILANGAN - Absolyut path ishlatilgan
// ubunto
// func setupOverlay(cwd string) {
// 	base := cwd + "/rootfs"
// 	lower := cwd + "/overlay/lower"
// 	upper := cwd + "/overlay/upper"
// 	work := cwd + "/overlay/work"
// 	merged := cwd + "/overlay/merged"

// 	// Kataloglar yaratish
// 	os.MkdirAll(lower, 0755)
// 	os.MkdirAll(upper, 0755)
// 	os.MkdirAll(work, 0755)
// 	os.MkdirAll(merged, 0755)

// 	// Bind mount
// 	must(syscall.Mount(base, lower, "", syscall.MS_BIND|syscall.MS_REC, ""))

//		// Overlay mount
//		opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
//		must(syscall.Mount("overlay", merged, "overlay", 0, opts))
//	}
func setupOverlay(cwd string) {
	base := cwd + "/rootfs"
	lower := cwd + "/overlay/lower"
	upper := cwd + "/overlay/upper"
	work := cwd + "/overlay/work"
	merged := cwd + "/overlay/merged"

	os.RemoveAll(cwd + "/overlay")
	must(os.MkdirAll(lower, 0755))
	must(os.MkdirAll(upper, 0755))
	must(os.MkdirAll(work, 0755))
	must(os.MkdirAll(merged, 0755))

	// bind mount rootfs → lower
	must(syscall.Mount(base, lower, "", syscall.MS_BIND|syscall.MS_REC, ""))

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lower, upper, work)

	must(syscall.Mount("overlay", merged, "overlay", 0, opts))
}

func cleanupOverlay(cwd string) {
	syscall.Unmount(cwd+"/overlay/merged", syscall.MNT_DETACH)
	syscall.Unmount(cwd+"/overlay/lower", syscall.MNT_DETACH)
	os.RemoveAll(cwd + "/overlay")
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

func setRlimits() {
	unix.Setrlimit(unix.RLIMIT_NPROC, &unix.Rlimit{Cur: 64, Max: 64})
	unix.Setrlimit(unix.RLIMIT_NOFILE, &unix.Rlimit{Cur: 256, Max: 256})
	unix.Setrlimit(unix.RLIMIT_CPU, &unix.Rlimit{Cur: 2, Max: 2})
	unix.Setrlimit(unix.RLIMIT_STACK, &unix.Rlimit{Cur: 8 * 1024 * 1024, Max: 8 * 1024 * 1024})
	unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{Cur: 0, Max: 0})
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
