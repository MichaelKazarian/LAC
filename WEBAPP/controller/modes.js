import { Communicator } from "./communicator.js";

const MODE            = "mode";
const MODE_MANUAL     = "mode-manual";
const MODE_ONCE_CYCLE = "mode-once-cycle";
const MODE_AUTO       = "mode-auto";


function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}


class Mode {
  _id;
  _description;
  _communicator
  _mainInterval
  _intervalVal = 50
  _inputState
  #task
  constructor() {
    this._description = "Base mode";
    this._id = MODE;
    this._communicator = new Communicator();
    this._inputState = {
      type: "input",
      degree: 51,
      rawinput: "[]",
      error: ""
    };

  }
  get description() {
    return this._description;
  }

  get id() {
    return this._id;
  }

  /**
   * Read the 3 registers starting at address 0 on device number this.#device
   * It is similar to executing
   * mbpoll -0 -m rtu -a 20 -b 9600 -t 4 -r 0 -c 3 -P none /dev/ttyUSB0
   * Note: time of readHoldingRegisters near 0.03 sec
   */
  _read = async () => {
    let d = this._communicator.device;
    await d.readHoldingRegisters(0, 3, async (err, data) => {
      if (data != null) {
        this._inputState.rawinput = (data.data);
        process.send(JSON.stringify(this._inputState));
      }
      else await sleep(50); // more than time of readHoldingRegisters to avoid crashes
    });
  }

  // TODO merge with activate
  run() {
    this._mainInterval = setInterval(() => {
      this.addTask(this._read);
    }, this._intervalVal);
  }

  
  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    this.run();   // TODO merge with run
    return true;
  }

  addTask(task) {
    this._communicator.addTask(task);
  }
  
  stop() {
    this._communicator.addTask("stop");
    clearInterval(this._mainInterval);
    console.log(`${this._description} stop`);
  }
}

//*****************************************************
class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
    this._id = MODE_MANUAL;
  }

}

class ModeOnceСycle extends Mode {
  constructor() {
    super();
    this._description = "Once cycle mode";
    this._id = MODE_ONCE_CYCLE;
  }  
}

class ModeAuto extends ModeOnceСycle {
  constructor() {
    super();
    this._description = "Automatic mode";
    this._id = MODE_AUTO;
  }
}

export {
  Mode,
  ModeManual,
  ModeOnceСycle,
  ModeAuto,
  MODE_MANUAL,
  MODE_ONCE_CYCLE,
  MODE_AUTO
};
