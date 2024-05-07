import { Controller } from "./controller.js";

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
});

// process.on('SIGINT', () => {
//     console.log(
//         `Child process terminated due to receipt of SIGINT`);
//     process.exit(0);
// });
