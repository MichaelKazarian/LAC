const ModbusRTU = require("modbus-serial");
var AsyncLock = require('async-lock');

var lock = new AsyncLock();

class Mode {
    
}

class ModeManual extends Mode {
    
}

class ModeAuto extends Mode {
    
}

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

class Communicator {
#inputState;
#outputState;
#mode;
#device
#que
    constructor(modbusId=20) {
        this.connectionInit(modbusId);
        this.#que = [];
        this.#mode = new ModeManual();
    }

    /**
     * Open connection to a serial port
     */
    connectionInit(modbusId) {
        this.#device = new ModbusRTU();
        this.#device.connectRTUBuffered("/dev/ttyUSB0", { baudRate: 9600 });
        this.#device.setTimeout(500);
        this.#device.setID(modbusId);
        this.#inputState = {
            degree: -1,
            rawinput: "[]",
            error: ""
        };
        this.#outputState = {
            r9: 1,
            r10: 1,
            r11: 1
        };
    }

    get inputState() {
        return this.#inputState;
    }

    /**
     * Read the 3 registers starting at address 0 on device number this.#device
     * It is similar to executing
     * mbpoll -0 -m rtu -a 20 -b 9600 -t 4 -r 0 -c 3 -P none /dev/ttyUSB0
     */
    read = async () => {
        await this.#device.readHoldingRegisters(0, 3, (err, data) => {
            this.#inputState.rawinput = (data.data);
        });
    }

    /**
     * write the values [0 ... 0xffff] to registers starting at address 3
     * It is similar to executing
     * mbpoll -0 -m rtu -a 20 -b 19200 -t 4 -r 3 -P none /dev/ttyUSB0 1 0 1
     */
    write = async (data) => {
        await this.#device.writeRegisters(3, data);
        console.log("B", new Date().getSeconds());
    }

    add = (f, pos="end") => {
        lock.acquire("key", () => {
            if (pos === "end") this.#que.push(f);
            else this.#que.unshift(f);
        });
    }

    do = () => {
        lock.acquire("key", () => {
	          return this.#que.shift();
        }).then((task) => {
            if (task === "read") {
                this.read();
                process.send(JSON.stringify(communicator.inputState));
            } else if (task.startsWith("radio&")) {
                let r = this.parseRadio(task);
                this.write(r);
            };
        })
            .catch(function(err) {
	              console.log("Err:", err.message); // output: error
            });
    }

    parseRadio = (msg) => {
        let res = [];
        let ids = msg.split("&")[1];
        for (let i in this.#outputState) {
            if (i === ids) {
                console.log("ID:", Math.abs(this.#outputState[i]-1));
                this.#outputState[i] = Math.abs(this.#outputState[i]-1);
            }
            res.push(this.#outputState[i]);
        }
        console.log(res);
        return res;
    }
}

let communicator = new Communicator();
// let manualState = false;

let allowInterval = true;
process.on('message', (msg) => {
    console.log('Message from parent:', msg, typeof(msg));
    if (msg === "STOP") {
        allowInterval = false;
    }
    if (msg === "read") {
        communicator.add("read");
    }
    if (msg.startsWith("radio&")) {
        communicator.add(msg);
        console.log("A", new Date().getSeconds());
    }

});

var i = setInterval(() => {
    if (!allowInterval) {
        clearInterval(i);
        process.exit();
    }
    communicator.add("read");
    communicator.do();
}, 50);

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
