"use client";

import { useEffect, useState } from "react";

export default function Home() {
  const [currentStep, setCurrentStep] = useState<
    "input" | "loading" | "result"
  >("input");

  useEffect(() => {
    if (currentStep !== "loading") {
      return;
    }

    const timerId = window.setTimeout(() => {
      setCurrentStep("result");
    }, 1500);

    return () => {
      window.clearTimeout(timerId);
    };
  }, [currentStep]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans text-zinc-900">
      <main className="flex w-full max-w-2xl flex-col gap-8 rounded-3xl bg-white px-10 py-16 shadow-xl">
        <div className="flex flex-col gap-2 text-center">
          <h1 className="font-semibold text-2xl tracking-tight">
            闇おみくじ（ステート骨格）
          </h1>
          <p className="text-sm text-zinc-500">入力 → 演出 → 結果</p>
        </div>

        {currentStep === "input" && (
          <section className="flex flex-col gap-4">
            <textarea
              className="min-h-[120px] w-full resize-none rounded-2xl border border-zinc-200 px-4 py-3 text-sm outline-none"
              placeholder="ここに闇を投げる（最大140字）"
            />
            <button
              className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white"
              type="button"
              onClick={() => setCurrentStep("loading")}
            >
              懺悔する
            </button>
          </section>
        )}

        {currentStep === "loading" && (
          <section className="flex flex-col items-center gap-4 text-center">
            <p className="font-medium text-base">
              少し待っていて、あなたのためのお告げを探すから。
            </p>
            <div className="h-2 w-full max-w-sm overflow-hidden rounded-full bg-zinc-100">
              <div className="h-full w-1/3 animate-pulse rounded-full bg-zinc-800" />
            </div>
          </section>
        )}

        {currentStep === "result" && (
          <section className="flex flex-col gap-4 text-center">
            <p className="font-medium text-base">
              今日の闇みくじ: ここに結果テキストが入ります。
            </p>
            <button
              className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
              type="button"
              onClick={() => setCurrentStep("input")}
            >
              もう一度懺悔する
            </button>
          </section>
        )}
      </main>
    </div>
  );
}
