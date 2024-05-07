const MODE            = 0;
const MODE_MANUAL     = 1;
const MODE_ONCE_CYCLE = 2;
const MODE_AUTO       = 3;

class Mode {
_description
  constructor() {
    this._description = "Base mode";
  }
  get description() {
    return this._description;
  }

  /**
   * Activate mode.
   * @return true if success; false otherwise;
   */
  activate() {
    return true;
  }
}

class ModeManual extends Mode {
  constructor() {
    super();
    this._description = "Manual mode";
  }

}

class ModeOnceСycle extends Mode {
  constructor() {
    super();
    this._description = "Once cycle mode";
  }  
}

class ModeAuto extends ModeOnceСycle {
  constructor() {
    super();
    this._description = "Automatic mode";
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
