const ModbusRTU = require("modbus-serial");
var AsyncLock = require('async-lock');

var lock = new AsyncLock();

class Mode {
    
}

class ModeManual extends Mode {
    
}

class ModeAuto extends Mode {
    
}

function sleep(ms) {
    return new Promise((resolve) => {
        setTimeout(resolve, ms);
    });
}


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
        this.#device.connectRTUBuffered("/dev/ttyUSB0", { baudRate: 9600, debug: true });
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
     * Note: time of readHoldingRegisters near 0.03 sec
     */
    read = async () => {
        await this.#device.readHoldingRegisters(0, 3, async (err, data) => {
            if (data != null) {
                this.#inputState.rawinput = (data.data);
            }
            else await sleep(50); // more than time of readHoldingRegisters to avoid crashes
        });
    }

    /**
     * write the values [0 ... 0xffff] to registers starting at address 3
     * It is similar to executing
     * mbpoll -0 -m rtu -a 20 -b 19200 -t 4 -r 3 -P none /dev/ttyUSB0 1 0 1
     * Note: Note: time of writeRegisters near 0.04 sec
     */
    write = async (data) => {
        await sleep(50); // more than time of writeRegisters to avoid crashes
        await this.#device.writeRegisters(3, data);
        // await sleep(50);
    }

    add = (f, pos="end") => {
            lock.acquire("key", () => {
                let lastIsSame = () => (this.#que.length > 0 && this.#que.at(-1) === f);
                // console.log("!@!", this.#que, !lastIsSame());
                if (!lastIsSame())
                    this.#que.push(f);
            }).then(this.do);
    }

    do = async () => {
        lock.acquire("key", async() => {
            while (this.#que.length > 0) {
                if (this.#que.length > 1) console.log("que0:", this.#que);
                try {
                    let task = this.#que.pop();
                    if (task === "read") {
                        await this.read();
                        process.send(JSON.stringify(this.#inputState));
                    } else if (task.startsWith("radio&")) {
                        let r = this.parseRadio(task);
                        await this.write(r);
                    };
                } catch (err) {
                    console.error("Err:", err.message); // output: error
                }
            }
        })
            .catch((e) => {
                console.error("ERRR: ", e.message);
            });
    }
    
    parseRadio = (msg) => {
        let res = [];
        let ids = msg.split("&")[1];
        for (let i in this.#outputState) {
            if (i === ids) {
                this.#outputState[i] = Math.abs(this.#outputState[i]-1);
            }
            res.push(this.#outputState[i]);
        }
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
    }

});

var i = setInterval(() => {
    if (!allowInterval) {
        clearInterval(i);
        process.exit();
    }
    communicator.add("read");
    // communicator.do();
}, 50);

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
