#rm -fr run
#./exe -logtostderr -conf www.newsmth.net.json
./cmd -log_dir="./log" -conf www.newsmth.net.json -wc 20
