function* operation() {
  let i;
  i = yield {type: "read", operation: "radio&r8"};
  console.log("Operation 1");
  while (i.degree <= 100) {
    // console.log("i:", i);
    if (parseInt(i.degree)%3 == 0) i = yield {type: "write", data: [0, 0, 0], operation: "radio&r7"};
    else i = yield {type: "write", data: [1, 1, 1], operation: "radio&r8"};
    i = yield {type: "read", operation: "radio&r8"};
  }
  return operation2;
}

function* operation2() {
  let i;
  i = yield {type: "read", operation: "radio&r10"};
  console.log("Operation 2");
  while (i.degree <= 200) {
    // console.log("i:", i);
    for (const item of i.rawinput) {
      if (item == 0)
        i = yield {type: "write", data: [1, 1, 1], operation: "radio&r10"};
    }
    i = yield {type: "read", operation: "radio&r10"};
  }
  return operation3;
};

function* operation3() {
  let i;
  i = yield {type: "read"};
  console.log("Operation 3");
  while (i.degree > 200) {
    if (i.degree<205) yield {type: "write", data: [0, 0, 0], operation: "radio&r8"};
    if (i.degree>205 && i.degree<250) {
      yield {type: "write", data: [1, 1, 1], operation: "radio&r7"};
    };
    if (i.degree>250 && i.degree<300) {
      yield {type: "write", data: [0, 1, 1], operation: "radio&r9"};
    }
    if (i.degree>300 && i.degree<500) {
      yield {type: "write", data: [0, 0, 1], operation: "radio&r10"};
    };
    if (i.degree>500 && i.degree<719) {
      yield {type: "write", data: [0, 0, 0], operation: "radio&r11"};
    };
    i = yield {type: "read", operation: "radio&r7"};
  }
  return;
}


class Operation1 {
#_id
  constructor() {
    this._id = "LAC base operation";
  }

  async do() {
    console.log(`${this._id} was executed`);
  }

  isOperationDone() {
    return true;
  }

  isOperationFinal() {
    return false;
  }

  nextOperation() {
    return none;
  }
}


export { operation }
