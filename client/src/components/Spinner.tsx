type SpinnerProps = {
  class?: string;
  ariaLabel?: string;
};

export function Spinner(props: SpinnerProps) {
  return (
    <img src="/swap/spinner.webp" alt={props.ariaLabel} class={props.class} />
  );
}
