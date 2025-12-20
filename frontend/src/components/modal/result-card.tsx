"use client";

type ResultCardProps = {
  resultText: string;
  onRetry: () => void;
  buttonLabel?: string;
};

export default function ResultCard({
  resultText,
  onRetry,
  buttonLabel = "もう一度懺悔する",
}: ResultCardProps) {
  return (
    <section className="flex flex-col items-center gap-6 text-center">
      <div className="w-full rounded-2xl border border-zinc-200 bg-zinc-50 px-5 py-4">
        <p className="font-medium text-base text-zinc-900 leading-relaxed whitespace-pre-wrap break-words">
          {resultText}
        </p>
      </div>
      <button
        className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
        type="button"
        onClick={onRetry}
      >
        {buttonLabel}
      </button>
    </section>
  );
}
