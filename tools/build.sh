cd ../cmd

echo "build start"
echo "build sentry_agent"
cd sentry_agent
go build

echo "build sentry_server"
cd ../sentry_server
go build

echo "build sentry_alarm"
cd ../sentry_alarm
go build


echo "build complete"

