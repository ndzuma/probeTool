.PHONY: probe install setup test clean

probe:
	go build -o probe ./cmd/probe

install: probe
	go install ./cmd/probe
	./probe setup

setup:
	./probe setup

test:
	cd ~/test-repo && probe --full

clean:
	rm -f probe
	rm -rf ~/.probe/probes/*.md
