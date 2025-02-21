import spinnerImg from "./spinner.webp";

type SpinnerProps = {
  class?: string;
  ariaLabel?: string;
};

export function Spinner(props: SpinnerProps) {
  return (
    <img
      src={spinnerImg.src}
      width={spinnerImg.width}
      height={spinnerImg.height}
      alt={props.ariaLabel}
      class={props.class}
    />
  );
}
