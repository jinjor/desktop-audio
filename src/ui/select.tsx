import React, { MouseEvent, useEffect, useState } from "react";
import { useHover } from "./util";

export const Select = (o: {
  list: string[];
  value: string;
  onChange: (value: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const callback = () => setOpen(false);
    window.addEventListener("click", callback);
    return () => window.removeEventListener("click", callback);
  }, []);
  const onClickTrigger = (e: MouseEvent) => {
    e.stopPropagation();
    setOpen(!open);
  };
  const onClickOption = (value: string) => {
    setOpen(false);
    if (value != o.value) {
      o.onChange(value);
    }
  };
  const size = 15;
  return (
    <div style={{ position: "relative" }}>
      <div
        style={{
          left: 0,
          right: 0,
          backgroundColor: "#444",
          border: "solid 1px #222",
          borderBottom: "solid 1px #333",
          padding: "1px 4px",
          whiteSpace: "nowrap",
        }}
        onClick={onClickTrigger}
      >
        <span>{o.value}</span>
        <svg
          style={{
            position: "absolute",
            right: "1px",
            width: `${size}px`,
            height: `${size}px`,
          }}
          viewBox={`-10 -10 20 20`}
        >
          <path
            // stroke="#eee"
            // strokeWidth={1}
            fill="#eee"
            d={`M -4 -2 h 8 l -4 4 Z`}
          />
        </svg>
      </div>
      <div
        style={{
          position: "absolute",
          minWidth: "100%",
          boxSizing: "border-box",
          display: open ? "block" : "none",
          backgroundColor: "#333",
          border: "solid 1px #222",
          zIndex: 1,
        }}
      >
        {o.list.map((item) => (
          <Option key={item} value={item} onClick={onClickOption} />
        ))}
      </div>
    </div>
  );
};

const Option = (o: { value: string; onClick: (value: string) => void }) => {
  const [hovered, listeners] = useHover();
  const onClick = (e: MouseEvent) => {
    e.stopPropagation();
    o.onClick(o.value);
  };
  return (
    <div
      {...listeners}
      style={{
        cursor: "pointer",
        backgroundColor: hovered ? "rgba(153,119,255,0.5)" : "transparent",
        padding: "1px 4px",
        whiteSpace: "nowrap",
      }}
      onClick={onClick}
    >
      {o.value}
    </div>
  );
};
