#!/bin/bash
dir=$(pwd)
cat "${dir}"/bin/inventory.txt > ~/.cminventory
echo "Inventory registered with cm"