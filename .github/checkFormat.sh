for file in $(find . -name "*.go") ; do
	diff -u <(echo -n) <(gofmt -d $file)
	if [ $? == "1" ] ; then
		exit 1
	fi
done
