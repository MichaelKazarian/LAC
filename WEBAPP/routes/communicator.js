const ModbusRTU = require("modbus-serial");

class Mode {
    
}

class ModeManual extends Mode {
    
}

class ModeAuto extends Mode {
    
}

class Communicator {
#inputState;
#mode;
#device
    constructor(modbusId=20) {
        this.connectionInit(modbusId);
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
    }

    get inputState() {
        return this.#inputState;
    }

    /**
     * Read the 3 registers starting at address 0 on device number this.#device
     * It is similar to executing
     * mbpoll -0 -m rtu -a 20 -b 9600 -t 4 -r 0 -c 3 -P none /dev/ttyUSB0
     */
    read = () => {
        this.#device.readHoldingRegisters(0, 3, (err, data) => {
            this.#inputState.rawinput = (data.data);
        });
    }

    /**
     * write the values [0 ... 0xffff] to registers starting at address 3
     * It is similar to executing
     * mbpoll -0 -m rtu -a 20 -b 19200 -t 4 -r 3 -P none /dev/ttyUSB0 1 0 1
     */
    write = (data) => {
        this.#device.writeRegisters(3, data);
        then(console.log);
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
        communicator.read();
    }
    if (msg === "radio1") {
        communicator.write([1, 1, 0]);
    }

});

var i = setInterval(() => {
    if (!allowInterval) {
        clearInterval(i);
        process.exit();
    }
    communicator.read();
    // console.log("XZ", communicator.inputState);
    process.send(JSON.stringify(communicator.inputState));
}, 100);

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
