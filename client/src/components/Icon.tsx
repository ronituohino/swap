type IconProps = {
  src: string;
  class?: string;
  ariaLabel?: string;
};

export function Icon(props: IconProps) {
  return (
    <img src={props.src} alt={props.ariaLabel || ""} class={props.class} />
  );
}
