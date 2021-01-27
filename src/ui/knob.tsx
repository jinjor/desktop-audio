import React, { useEffect, useState } from "react";

const Tooptip = (o: {
  text: string;
  position: { x: number; y: number } | null;
}) => {
  if (o.position == null) {
    return null;
  }
  return (
    <div
      style={{
        position: "fixed",
        left: `${o.position.x + 10}px`,
        top: `${o.position.y + 15}px`,
        border: "solid 1px #aaa",
        backgroundColor: "#444",
        zIndex: 1,
        padding: "1px 2px",
      }}
    >
      {o.text}
    </div>
  );
};

type LabeledKnobOptions = KnobOptions & {
  label: string;
};

export const LabeledKnob = (o: LabeledKnobOptions) => {
  const { label, ...knobOptions } = o;
  return (
    <div style={{ display: "flex", flexFlow: "column", textAlign: "center" }}>
      <Knob {...knobOptions} />
      <label style={{ fontSize: "12px" }}>{label}</label>
    </div>
  );
};

type KnobOptions = {
  exponential: boolean;
  min: number;
  max: number;
  value: number;
  steps: number | null;
  onChange: (value: number) => void;
};

export const Knob = (o: KnobOptions) => {
  const [mousePosition, setMousePosition] = useState<{
    x: number;
    y: number;
  } | null>(null);
  const size = 40;
  const v = (o.value - o.min) / (o.max - o.min);
  const onInput = (v: number) => {
    o.onChange(o.min + (o.max - o.min) * v);
  };
  return (
    <KnobHandler
      value={v}
      steps={o.steps}
      onInput={onInput}
      onHold={setMousePosition}
      style={{
        display: "inline-block",
        width: `${size}px`,
        height: `${size}px`,
        position: "relative",
        userSelect: "none",
      }}
    >
      <KnobView size={size} value={v} />
      <Tooptip text={o.value.toFixed(1)} position={mousePosition} />
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
          top: "0",
          left: "0",
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
  onHold: (
    data: {
      x: number;
      y: number;
    } | null
  ) => void;
  onInput: (value: number) => void;
}) => {
  const { value, steps, children, onHold, onInput, ...props } = o;
  const valuePerX = 1 / 100;
  const valuePerY = 1 / 100;
  const [start, setStart] = useState<{
    x: number;
    y: number;
    value: number;
  } | null>(null);
  const onMouseDown = (e: React.MouseEvent) => {
    const x = e.clientX;
    const y = e.clientY;
    setStart({ x, y, value });
    onHold({ x, y });
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
      onHold({ x, y });
      if (v !== value) {
        onInput(v);
      }
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
      onHold(null);
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
