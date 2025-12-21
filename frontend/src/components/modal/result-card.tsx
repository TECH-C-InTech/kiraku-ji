"use client";

type ResultCardProps = {
  resultText: string;
  onRetry: () => void;
  buttonLabel?: string;
  buttonClassName?: string;
  secondaryButtonLabel?: string;
  onSecondary?: () => void;
  secondaryButtonClassName?: string;
  reverseButtons?: boolean;
};

/** おみくじ結果と操作ボタンをまとめて表示する。 */
export default function ResultCard({
  resultText,
  onRetry,
  buttonLabel = "もう一度懺悔する",
  buttonClassName = "border-zinc-300 text-zinc-700",
  secondaryButtonLabel,
  onSecondary,
  secondaryButtonClassName = "border-zinc-300 text-zinc-700",
  reverseButtons = false,
}: ResultCardProps) {
  const mergedButtonClassName = `rounded-xl border px-6 py-3 font-semibold text-base ${buttonClassName}`;
  const mergedSecondaryButtonClassName = `rounded-xl border px-6 py-3 font-semibold text-base ${secondaryButtonClassName}`;
  const hasSecondaryButton =
    typeof onSecondary === "function" && Boolean(secondaryButtonLabel);
  const actionContainerClassName = hasSecondaryButton
    ? "flex flex-row flex-wrap items-center justify-center gap-3"
    : "flex flex-col items-center";
  const primaryButton = (
    <button className={mergedButtonClassName} type="button" onClick={onRetry}>
      {buttonLabel}
    </button>
  );
  const secondaryButton = hasSecondaryButton ? (
    <button
      className={mergedSecondaryButtonClassName}
      type="button"
      onClick={onSecondary}
    >
      {secondaryButtonLabel}
    </button>
  ) : null;

  return (
    <section className="flex flex-col items-center gap-6 text-center">
      <div className="relative w-full rounded-lg bg-zinc-900 px-5 py-4">
        <svg
          className="pointer-events-none absolute inset-0 h-full w-full text-zinc-500"
          viewBox="0 0 100 100"
          preserveAspectRatio="none"
          aria-hidden="true"
        >
          <rect
            x="1"
            y="1"
            width="98"
            height="98"
            rx="6"
            ry="6"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeDasharray="14 6 4 10 8 12 6 5 9 7"
          />
        </svg>
        <p className="relative z-10 whitespace-pre-wrap break-words font-medium text-base text-zinc-100 leading-relaxed">
          {resultText}
        </p>
      </div>
      <div className={actionContainerClassName}>
        {reverseButtons ? secondaryButton : primaryButton}
        {reverseButtons ? primaryButton : secondaryButton}
      </div>
    </section>
  );
}
