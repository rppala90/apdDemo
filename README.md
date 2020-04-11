Credit to : https://github.com/uber-common/cadence-samples

Under Cadence:

git clone git@github.com:uber/cadence.git

docker-compose up

./cadence --domain apd-domain domain register

Under apdDemo:

make bins

./bin/apdserver

./bin/apddemo -m worker
