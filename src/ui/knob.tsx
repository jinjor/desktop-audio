import React, { useEffect, useState } from "react";

export const Knob = (o: {
  exponential: boolean;
  min: number;
  max: number;
  value: number;
  steps: number | null;
  onInput: (value: number) => void;
}) => {
  const size = 40;
  const v = (o.value - o.min) / (o.max - o.min);
  const onInput = (v: number) => {
    o.onInput(o.min + (o.max - o.min) * v);
  };
  return (
    <KnobHandler
      value={v}
      steps={o.steps}
      onInput={onInput}
      style={{
        width: `${size}px`,
        height: `${size}px`,
        position: "relative",
        userSelect: "none",
      }}
    >
      <KnobView size={size} value={v} />
    </KnobHandler>
  );
};

const KnobView = (o: { size: number; value: number }) => {
  const { size, value } = o;
  const knobR = 15;
  const knobSize = knobR * 2;
  const knobOffset = (size - knobR * 2) / 2;
  const pointD = 11;
  const pointR = 2;
  const startRad = Math.PI * (-4 / 3);
  const endRad = Math.PI * (1 / 3);
  const valueRad = value * endRad + (1 - value) * startRad;
  const slitWidth = 2;
  const slitR = knobR + 2 + slitWidth / 2;
  const startX = slitR * Math.cos(startRad);
  const startY = slitR * Math.sin(startRad);
  const endX = slitR * Math.cos(endRad);
  const endY = slitR * Math.sin(endRad);
  const valueX = slitR * Math.cos(valueRad);
  const valueY = slitR * Math.sin(valueRad);
  const pointX = pointD * Math.cos(valueRad);
  const pointY = pointD * Math.sin(valueRad);
  const slitColor = "#97f";
  const largeArc = valueRad - startRad >= Math.PI ? 1 : 0;
  if (slitR * 2 + slitWidth > size) {
    throw new Error("assertion error");
  }
  const Path = (o: {
    endX: number;
    endY: number;
    largeArc: number;
    color: string;
  }) => (
    <path
      stroke={o.color}
      strokeWidth={slitWidth}
      fill="none"
      d={`M ${startX} ${startY} A ${slitR} ${slitR} 0 ${o.largeArc} 1 ${o.endX} ${o.endY}`}
    />
  );
  return (
    <React.Fragment>
      <div
        style={{
          position: "absolute",
          top: `${knobOffset}px`,
          left: `${knobOffset}px`,
          width: `${knobSize}px`,
          height: `${knobSize}px`,
          borderRadius: "50%",
          boxShadow: "0 1px 2px 1px rgba(0,0,0,0.4)",
        }}
      />
      <svg
        style={{
          position: "absolute",
          width: `${size}px`,
          height: `${size}px`,
        }}
        viewBox={`${-size / 2} ${-size / 2} ${size} ${size}`}
      >
        <Path endX={endX} endY={endY} largeArc={1} color="#000" />
        <Path
          endX={valueX}
          endY={valueY}
          largeArc={largeArc}
          color={slitColor}
        />
        <circle r={pointR} cx={pointX} cy={pointY} fill={slitColor} />
      </svg>
    </React.Fragment>
  );
};

const KnobHandler = (o: {
  [key: string]: any;
  value: number;
  steps: number | null;
  children: any;
  onInput: (value: number) => void;
}) => {
  const { value, steps, children, onInput, ...props } = o;
  const valuePerX = 1 / 200;
  const valuePerY = 1 / 200;
  const [start, setStart] = useState<{
    x: number;
    y: number;
    value: number;
  } | null>(null);
  const onMouseDown = (e: React.MouseEvent) => {
    setStart({ x: e.clientX, y: e.clientY, value });
  };
  useEffect(() => {
    const onMouseMove = (e: MouseEvent) => {
      if (start == null) {
        return;
      }
      const x = e.clientX;
      const y = e.clientY;
      const dv = (x - start.x) * valuePerX + (start.y - y) * valuePerY;
      let v = Math.min(1, Math.max(0, start.value + dv));
      if (steps != null) {
        v = Math.floor(v * steps) / steps;
      }
      if (v === value) {
        return;
      }
      onInput(v);
    };
    window.addEventListener("mousemove", onMouseMove);
    return () => window.removeEventListener("mousemove", onMouseMove);
  });
  useEffect(() => {
    const onMouseUp = (e: MouseEvent) => {
      if (start == null) {
        return;
      }
      setStart(null);
    };
    window.addEventListener("mouseup", onMouseUp);
    return () => window.removeEventListener("mouseup", onMouseUp);
  });
  return (
    <div onMouseDown={onMouseDown} {...props}>
      {children}
    </div>
  );
};
