import React from "react";

const wrap = (size: number, path: JSX.Element) => {
  return (
    <svg
      style={{ verticalAlign: "middle" }}
      width={size}
      height={size}
      viewBox="-10 -10 20 20"
    >
      {path}
    </svg>
  );
};

export const sine = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,0 Q -4 -12 0 0 T 8 0"
      fill="none"
      stroke="white"
      stroke-width="1.8"
    />
  );
export const triangle = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,0 L -4,-7 4,7 8,0"
      fill="none"
      stroke="white"
      stroke-width="1.5"
    />
  );
export const square = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,0 L -8,-6 0,-6 0,6 8,6 8,0"
      fill="none"
      stroke="white"
      stroke-width="1.5"
    />
  );
export const pulse = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,0 L -8,-6 -4,-6 -4,6 8,6 8,0"
      fill="none"
      stroke="white"
      stroke-width="1.5"
    />
  );
export const saw = (size: number) =>
  wrap(
    size,
    <path d="M -8,6 L 8,-6 8,6" fill="none" stroke="white" stroke-width="1.5" />
  );
export const sawRev = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,6 L -8,-6 8,6"
      fill="none"
      stroke="white"
      stroke-width="1.5"
    />
  );
export const noise = (size: number) =>
  wrap(
    size,
    <path
      d="M -8,0 L -6,-2 -4,6 -2,-8 0,5 2,-5 4,7 6,-4 8,2"
      fill="none"
      stroke="white"
      stroke-width="1.5"
    />
  );
