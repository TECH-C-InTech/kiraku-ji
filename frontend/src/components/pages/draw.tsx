"use client";

type FortuneDrawProps = {
  message?: string;
};

export default function FortuneDrawPage({
  message = "少し待っていて、あなたのためのお告げを探すから。",
}: FortuneDrawProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans text-zinc-900">
      <main className="flex w-full max-w-xl flex-col gap-6 rounded-3xl bg-white px-8 py-12 text-center shadow-lg">
        <div className="flex h-40 w-40 items-center justify-center self-center rounded-full border border-zinc-300 bg-zinc-100 text-xs text-zinc-500">
          黒グラキャラ（仮）
        </div>
        <p className="font-medium text-base">{message}</p>
        <div
          className="h-2 w-full overflow-hidden rounded-full bg-zinc-200"
          role="progressbar"
          aria-busy="true"
          aria-label="きらくじを引いています"
        >
          <div className="h-full w-1/3 animate-pulse rounded-full bg-zinc-700" />
        </div>
        <p className="text-sm text-zinc-500">きらくじを引いています...</p>
      </main>
    </div>
  );
}
