#! /bin/sh
### BEGIN INIT INFO
# Provides:          teamwork-server.sh
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Runs the TeamWorkServer binary
# Description:       Makes sure the TeamWorkServer binary starts on boot
### END INIT INFO

case "$1" in
  start)
    echo "Starting TeamWorkServer"
    /opt/src/github.com/Banrai/TeamWork.io/server/TeamWorkServer -dbName=teamworkdb \
      -dbUser=teamworkio \
      -dbPass=dbPassGoesHere \
      -dbSSL=true \
      -host=127.0.0.1 \
      -extPort=8001 \
      -port=8001 \
      -ssl=true \
      -templates=/opt/src/github.com/Banrai/TeamWork.io/html/templates \
      -stripePK=pk_live_GoesHere \
      -stripeSK=sk_live_GoesHere >> /opt/TeamWorkServer.log 2>&1 &
    ;;
  stop)
    echo "Stopping TeamWorkServer"
    killall TeamWorkServer
    ;;
  *)
    echo "Usage: /etc/init.d/teamwork-server.sh {start|stop}"
    exit 1
    ;;
esac

exit 0
