.PHONY: agent probe install test clean

agent:
	cd agent && npm install

probe: agent
	go build -o probe ./cmd/probe

install: agent
	go install ./cmd/probe

test:
	cd ~/test-repo && probe --full

clean:
	rm -f probe
	rm -f probes/*.md
