import React from "react";

export const Knob = (o: {
  exponential: boolean;
  min: number;
  max: number;
  value: number;
  step: number;
  onInput: (value: number) => void;
}) => {
  const size = 40;
  const knobR = 15;
  const knobSize = knobR * 2;
  const knobOffset = (size - knobR * 2) / 2;
  const pointD = 11;
  const pointR = 2;
  const v = o.value / (o.max - o.min);
  const startRad = Math.PI * (-4 / 3);
  const endRad = Math.PI * (1 / 3);
  const valueRad = v * endRad + (1 - v) * startRad;
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
    ></path>
  );
  return (
    <div
      style={{
        width: `${size}px`,
        height: `${size}px`,
        position: "relative",
      }}
    >
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
      ></div>
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
    </div>
  );
};
