ssh server2 rm /root/dev/go-learn/free-class/*
scp .\* server2:/root/dev/go-learn/free-class/
# ssh server2 "docker exec nginx nginx -s reload"