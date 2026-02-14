#!/bin/bash
# diagnose.sh

echo "🔍 OverlayFS diagnostikasi..."
echo "================================"

# 1. Kernel moduli
echo -n "1. Overlay moduli: "
if lsmod | grep -q overlay; then
    echo "✅ Yuklangan"
else
    echo "❌ Yuklanmagan"
    echo "   Yuklanmoqda: sudo modprobe overlay"
    sudo modprobe overlay
fi

# 2. Filesystem support
echo -n "2. Overlay filesystem: "
if cat /proc/filesystems | grep -q overlay; then
    echo "✅ Qo'llab-quvvatlanadi"
else
    echo "❌ Qo'llab-quvvatlanmaydi"
fi

# 3. Test mount
echo "3. Test overlay mount:"
mkdir -p /tmp/test/{lower,upper,work,merged}
echo "test" > /tmp/test/lower/test.txt

sudo mount -t overlay overlay \
    -o lowerdir=/tmp/test/lower,upperdir=/tmp/test/upper,workdir=/tmp/test/work \
    /tmp/test/merged

if [ $? -eq 0 ]; then
    echo "   ✅ Mount muvaffaqiyatli"
    cat /tmp/test/merged/test.txt
    sudo umount /tmp/test/merged
else
    echo "   ❌ Mount xatosi"
fi

rm -rf /tmp/test
echo "================================"
