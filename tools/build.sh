#/bin/sh

ToolsRelativePath=`dirname $0`
ToolsPath=`cd $ToolsRelativePath; pwd`
BuildTargets="sentry_agent sentry_server sentry_alarm"

# build the executable targets
function build() {
	echo "build start"

	cd ../cmd
	for target in $BuildTargets
	do 
		echo "build $target"
		cd $target
		go build
		cd ..
	done
	
	echo "build complete"
}

function clean_old_pkg() {
	echo "clean old packages if exist"
	cd $ToolsPath

	for target in $BuildTargets
    do
		if [ -e $target ]; then
			rm -fr $target
		fi

		if [ -e $target.tar.gz ]; then
        	rm -fr $target.tar.gz
    	fi
	done

	echo "clean old packages complete"
}

function package() {
	echo "make packages"

	for target in $BuildTargets
    do
		mkdir $target
		cp ../cmd/$target/$target ./$target/
		if [ $target == 'sentry_agent' ]; then
			cp ../configs/SentryAgent.conf ./$target/
			cp -fr ../scripts ./$target/
		elif [ $target == 'sentry_server' ]; then
			cp ../configs/SentryServer.conf ./$target/
			cp -fr ../frontend ./$target/
		else 
			cp ../configs/SentryAlarm.conf ./$target/
		fi

		tar zcvf $target.tar.gz $target
	done
}

function clean_dir() {
	echo "clean immediately directory"

    for target in $BuildTargets
    do
		rm -fr $target
	done
}

build
clean_old_pkg
package
clean_dir

