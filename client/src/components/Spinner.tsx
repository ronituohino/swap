type SpinnerProps = {
  class?: string;
  ariaLabel?: string;
};

export function Spinner(props: SpinnerProps) {
  return (
    <img src="/public/spinner.webp" alt={props.ariaLabel} class={props.class} />
  );
}
