type SpinnerProps = {
  class?: string;
};

export function Spinner(props: SpinnerProps) {
  return <img src="/public/spinner.webp" alt="" class={props.class} />;
}
