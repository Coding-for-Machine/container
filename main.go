package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	// docker run [OPTION] <image> [args]
	// 1. Foydalanuvchi yozgan buyruqni tekshiramiz (masalan: go run main.go run )
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Buyruq topilmadi!")
	}
}

func run() {
	fmt.Println("Konteyner ishga tushmoqda...")
	// 2. Foydalanuvchi yuborgan qolgan barcha argumentlarni chop etamiz
	fmt.Printf("Bajariladigan buyruqlar: %v\n", os.Args[2:])
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}
	must(cmd.Run())

}

func child() {
	must(syscall.Sethostname([]byte("container")))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
