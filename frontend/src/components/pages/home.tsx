"use client";

import { type ChangeEvent, useState } from "react";
import { fetchRandomDraw } from "@/lib/draws";
import { createPost } from "@/lib/posts";

type Step = "input" | "loading" | "ready" | "result" | "error";

export default function HomePage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [currentStep, setCurrentStep] = useState<Step>("input");
  const [content, setContent] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [resultText, setResultText] = useState("");
  const defaultPostError = "投稿に失敗しました";
  const defaultDrawError = "おみくじの取得に失敗しました";
  const contentLength = content.length;
  const trimmedLength = content.trim().length;
  const isSubmitDisabled = trimmedLength === 0 || contentLength > 140;
  const isSubmitButtonDisabled = isSubmitDisabled || currentStep === "loading";

  const handleContentChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    setContent(event.target.value);
  };

  const handleRetry = (options?: { clearContent?: boolean }) => {
    if (options?.clearContent) {
      setContent("");
      setResultText("");
    }
    setErrorMessage("");
    setCurrentStep("input");
  };

  const handleSubmit = async () => {
    if (isSubmitDisabled || currentStep === "loading") {
      return;
    }
    setCurrentStep("loading");
    setErrorMessage("");
    try {
      await createPost(content.trim());
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : defaultPostError,
      );
      setCurrentStep("error");
      return;
    }

    setCurrentStep("ready");
  };

  const handleDraw = async () => {
    if (currentStep !== "ready") {
      return;
    }
    setCurrentStep("loading");
    setErrorMessage("");
    try {
      const draw = await fetchRandomDraw();
      setResultText(draw.result);
      setCurrentStep("result");
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : defaultDrawError,
      );
      setCurrentStep("error");
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 px-4 py-8 font-sans text-zinc-900 md:px-0">
      <div className="flex w-full max-w-xl flex-col items-center gap-4">
        <h1 className="font-semibold text-xl">きらくじ（仮UI）</h1>
        <button
          className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white transition hover:bg-zinc-800"
          type="button"
          onClick={() => {
            setIsModalOpen(true);
            handleRetry({ clearContent: true });
          }}
        >
          闇を投げる
        </button>
      </div>

      {isModalOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4 py-8"
          role="dialog"
          aria-modal="true"
          onClick={() => {
            setIsModalOpen(false);
            handleRetry();
          }}
        >
          <main
            className="relative flex w-full max-w-lg flex-col gap-8 rounded-3xl bg-white px-6 py-10 shadow-lg md:max-w-xl md:px-8 md:py-12"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              className="absolute right-4 top-4 rounded-full bg-zinc-100 px-3 py-1 text-xs font-semibold text-zinc-600 hover:bg-zinc-200"
              type="button"
              onClick={() => {
                setIsModalOpen(false);
                handleRetry();
              }}
            >
              閉じる
            </button>

            {currentStep === "input" && (
              <section className="flex flex-col gap-4">
                <textarea
                  className="min-h-[140px] w-full resize-none rounded-2xl border border-zinc-200 px-4 py-3 text-sm outline-none"
                  maxLength={140}
                  placeholder="ここに闇を投げる（最大140字）"
                  value={content}
                  onChange={handleContentChange}
                />
                <div className="text-right text-xs text-zinc-500">
                  {contentLength}/140
                </div>
                <button
                  className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white disabled:cursor-not-allowed disabled:bg-zinc-400"
                  type="button"
                  onClick={handleSubmit}
                  disabled={isSubmitButtonDisabled}
                >
                  懺悔する
                </button>
              </section>
            )}

            {currentStep === "loading" && (
              <section className="flex flex-col items-center gap-4 text-center">
                <div className="flex h-40 w-40 items-center justify-center rounded-full border border-zinc-300 bg-zinc-100 text-xs text-zinc-500">
                  黒グラキャラ（仮）
                </div>
                <p className="font-medium text-base">
                  少し待っていて、あなたのためのお告げを探すから。
                </p>
                <div
                  className="h-2 w-full max-w-sm overflow-hidden rounded-full bg-zinc-200"
                  role="progressbar"
                  aria-busy="true"
                  aria-label="きらくじを引いています"
                >
                  <div className="h-full w-1/3 animate-pulse rounded-full bg-zinc-700" />
                </div>
                <p className="text-sm text-zinc-500">
                  きらくじを引いています...
                </p>
              </section>
            )}

            {currentStep === "ready" && (
              <section className="flex flex-col gap-4 text-center">
                <p className="font-medium text-base">
                  投稿が完了しました。おみくじを引きますか？
                </p>
                <button
                  className="rounded-full bg-zinc-900 px-6 py-3 font-semibold text-sm text-white"
                  type="button"
                  onClick={handleDraw}
                >
                  きらくじを引く
                </button>
              </section>
            )}

            {currentStep === "result" && (
              <section className="flex flex-col gap-4 text-center">
                <p className="font-medium text-base">{resultText}</p>
                <button
                  className="rounded-full border border-zinc-300 px-6 py-3 font-semibold text-sm text-zinc-700"
                  type="button"
                  onClick={() => handleRetry({ clearContent: true })}
                >
                  もう一度懺悔する
                </button>
              </section>
            )}

            {currentStep === "error" && (
              <section className="flex flex-col gap-4 text-center">
                <p className="font-medium text-base text-red-600">
                  {errorMessage}
                </p>
                <button
                  className="rounded-full border border-red-200 px-6 py-3 font-semibold text-red-700 text-sm"
                  type="button"
                  onClick={() => handleRetry()}
                >
                  入力画面へ戻る
                </button>
              </section>
            )}
          </main>
        </div>
      )}
    </div>
  );
}
