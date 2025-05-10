# Dualnet Chat

Dualnet Chat is a demonstration of a command-line chat server and client applications using the OSI Layer 4 (Transport Layer) Transmission Control Protocol (TCP) and User Datagram Protocol (UDP). The technical differences are compared and contrasted.

## Deliverables

As the final project for the CMPS2242 Systems Programming & Computer Organization course at the University of Belize, below are the deliverables prepared for submission:

- Demonstration/reflections video ([YouTube]())
- Presentation slides ([Google Slides](https://docs.google.com/presentation/d/1PG-6aN_arnryI0kGtF-N80mj_TJANbFhm8J4itljvZg/edit?usp=sharing))
- Plan of action document ([Google Docs](https://docs.google.com/document/d/1DezkvbLiQQy_j1XP9Z_B_D_w3c2mnrVemJdGQyrfEuw/edit?usp=sharing))

## Building the Project

### Prerequisites

- `go` 1.24.2 or later
- `make` (build tool)
- `tc` (traffic control) (optional for network tests)
- `pmap` (process memory map) (optional for memory tests)

### Steps

1. Clone the repository:

```
git clone https://github.com/jennxsierra/dualnet-chat.git
```

2. Change directory to the project folder:

```
cd dualnet-chat
```

3. Build the project using `make`:

```
make
```

4. Run a server application in one terminal window and the respective client application in another. They should be built in the `bin` folder. From the project root, you can use a set of the following:

> [!TIP]
> The default port for the TCP server is 4000, and the default port for the UDP server is 4001. You can specify a different port with the `--port` flag. For example, `./bin/tcp-server --port 4040`.

- `./bin/tcp-server`
- `./bin/tcp-client`
- `./bin/udp-server`
- `./bin/udp-client`

## Tests

### Network Tests

Several network tests are included in the `tests` folder. They are run using custom makefile rules, and results are logged into a `results` folder. The following are the available tests and the defined network conditions. For more information, examine the `Makefile` in the project root.

- `make test-tcp-network-good`
- `make test-tcp-network-normal`
- `make test-tcp-network-bad`
- `make test-udp-network-good`
- `make test-udp-network-normal`
- `make test-udp-network-bad`

```
# Network Conditions

# good network (fast and stable)
GOOD_LATENCY = 10ms
GOOD_LOSS = 0.1%
GOOD_RATE = 10mbit

# normal network (moderate)
NORMAL_LATENCY = 50ms
NORMAL_LOSS = 1%
NORMAL_RATE = 5mbit

# bad network (slow and unreliable)
BAD_LATENCY = 200ms
BAD_LOSS = 5%
BAD_RATE = 1mbit
```

### Memory Tests

To test the memory consumption of the server and client applications, run one in a terminal window, take note of the port it is running on, and then run the following script with the port number as an argument.

> [!NOTE]
> You may need to make the script executable first. You can do this with the command `chmod +x scripts/check_memory.sh`.

```
./scripts/check_memory.sh <port>
```

## Project Structure Highlights

- `cmd` directory contains the `main.go` files for the respective server and client applications. They are simple and only handle the `--port` flags.
- `internal` directory contains the core logic of the server and client applications. The `server.go` and `client.go` files utilize a struct with defined methods to handle the TCP and UDP protocols.
- `scripts` and `tests` directories contain code for application testing.

## Cleanup

To remove the build artifacts and test results, run the following command:

```
make clean
```
