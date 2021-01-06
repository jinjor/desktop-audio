import React, { useState } from "react";
import { ipcRenderer } from "electron";

const Note: React.FC<{
  number: number;
  pressed: boolean;
}> = ({ children, number, pressed }) => {
  const onMouseDown = () => {
    ipcRenderer.send("audio", ["note_on", String(number)]);
  };
  const onMouseUp = () => {
    ipcRenderer.send("audio", ["note_off"]);
  };
  const onMouseEnter = () => {
    if (pressed) {
      ipcRenderer.send("audio", ["note_on", String(number)]);
    }
  };
  const onMouseLeave = () => {
    if (pressed) {
      ipcRenderer.send("audio", ["note_off"]);
    }
  };
  return (
    <button
      onMouseDown={onMouseDown}
      onMouseUp={onMouseUp}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      {children}
    </button>
  );
};

export const Notes = () => {
  const [pressed, setPressed] = useState(false);
  const onMouseDown = () => {
    setPressed(true);
  };
  const onMouseUp = () => {
    setPressed(false);
  };
  const onMouseLeave = () => {
    setPressed(false);
  };
  return (
    <div
      onMouseDown={onMouseDown}
      onMouseUp={onMouseUp}
      onMouseLeave={onMouseLeave}
    >
      <Note number={60} {...{ pressed }}>
        C
      </Note>
      <Note number={62} {...{ pressed }}>
        D
      </Note>
      <Note number={64} {...{ pressed }}>
        E
      </Note>
    </div>
  );
};
