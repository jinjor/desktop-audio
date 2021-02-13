import net from "net";
import readline from "readline";
import { spawn } from "child_process";
import { existsSync, unlinkSync } from "fs";

const sockPath = "/tmp/desktop-audio.sock";

export class AudioClient {
  private client?: net.Socket;
  private connected = true;
  onMessage?: (command: string[]) => void;
  onConnected?: () => void;
  onDisconnected?: () => void;
  onError?: (err: Error) => void;
  constructor() {}
  async connect(): Promise<void> {
    const p = spawn("./dist/audio", [], {
      stdio: "inherit",
    });
    const interval = 200;
    const maxRetries = 20;
    for (let i = 0; ; i++) {
      if (existsSync(sockPath)) {
        break;
      }
      console.log(`retrying...`);
      if (i >= maxRetries) {
        throw new Error(
          "could not connect to the audio server: sock file not found"
        );
      }
      await new Promise((resolve) => setTimeout(resolve, interval));
    }
    this.client = new net.Socket();
    this.client.on("connect", () => {
      this.connected = true;
      this.onConnected?.();
    });
    this.client.on("end", () => {
      this.connected = false;
      this.onDisconnected?.();
    });
    this.client.on("error", (e: Error) => {
      p.kill("SIGINT");
      unlinkSync(sockPath);
      this.onError?.(e);
    });
    const rl = readline.createInterface({
      input: this.client,
      crlfDelay: Infinity,
    });
    rl.on("line", (line) => {
      if (this.onMessage != null) {
        const command = line.split(/\s+/).map(decodeURIComponent);
        this.onMessage(command);
      }
    });
    this.client.connect(sockPath);
  }
  send(command: string[]) {
    if (this.client == null || !this.connected) {
      throw new Error("connection is closed");
    }
    this.client.write(`${command.map(encodeURIComponent).join(" ")}\n`);
  }
}
