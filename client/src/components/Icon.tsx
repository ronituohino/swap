type IconProps = {
  class?: string;
};

export function Icon(props: IconProps) {
  return <img src="/public/favicon.ico" alt="" class={props.class} />;
}
