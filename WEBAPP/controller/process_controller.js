import { Controller } from "./controller.js";
import {
  MODE_MANUAL, MODE_ONCE_CYCLE, MODE_AUTO
} from "./modes.js";

let controller = new Controller();
controller.run();

process.on('message', (msg) => {
  console.log('Message from parent:', msg, typeof(msg));
  if (msg === "STOP") {
    controller.stop();
  }
  if (msg === "read") {
    controller.addTask("read");
  }
  if (msg.startsWith("radio&")) {
    controller.addTask(msg);
  }
  if (msg === "mode-manual") {
    controller.setMode(MODE_MANUAL)
      .then(() => process.send("mode-set:success"))
      .catch(() => process.send("mode-set:error"));
  }
  if (msg === "mode-once-cycle") {
    controller.setMode(MODE_ONCE_CYCLE)
      .then(() => process.send("mode-set:success"))
      .catch(() => process.send("mode-set:error"));
  }
  if (msg === "mode-auto") {
    controller.setMode(MODE_AUTO)
      .then(() => process.send("mode-set:success"))
      .catch(() => process.send("mode-set:error"));
  }
});

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
