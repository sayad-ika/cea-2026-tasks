import { useState, useCallback } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

import { applicationSchema, type ApplicationFormData } from "./lib/schema";
import { submitApplication } from "./lib/mockApi";

import { PersonalInfo } from "./components/PersonalInfo";
import { Experience } from "./components/Experience";
import { ReviewSubmit } from "./components/ReviewSubmit";
import { SubmissionSuccess } from "./components/SubmissionSuccess";
import Toast from "./components/Toast";
import type { ToastType } from "./type/toast";
import logoUrl from "./assets/new_vite.svg";

const STEP_LABELS = ["Personal Info", "Experience", "Review & Submit"];

const STEP_FIELDS: Array<Array<keyof ApplicationFormData>> = [
    [
        "firstName",
        "lastName",
        "email",
        "phone",
        "linkedinUrl",
        "githubUrl",
        "portfolioUrl",
    ],
    ["hasExperience", "yearsOfExperience", "previousJobs", "resume"],
    [],
];

export default function App() {
    const [step, setStep] = useState(0);
    const [submitting, setSubmitting] = useState(false);
    const [submittedData, setSubmittedData] =
        useState<ApplicationFormData | null>(null);
    const [toast, setToast] = useState<{ message: string; type: ToastType }>({
        message: "",
        type: null,
    });

    const methods = useForm<ApplicationFormData>({
        resolver: zodResolver(applicationSchema),
        defaultValues: {
            firstName: "",
            lastName: "",
            email: "",
            phone: "",
            linkedinUrl: "",
            githubUrl: "",
            portfolioUrl: "",
            hasExperience: false,
            yearsOfExperience: 0,
            previousJobs: [],
            resume: undefined,
        },
        mode: "onBlur",
        reValidateMode: "onChange",
    });

    const {
        handleSubmit,
        trigger,
        formState: { isValidating },
    } = methods;

    async function handleNext(e?: React.MouseEvent) {
        e?.preventDefault();

        const fields = STEP_FIELDS[step];
        const valid = await trigger(fields);

        if (!valid) {
            const currentErrors = methods.formState.errors;
            const firstErrorField = fields.find((f) => currentErrors[f]);
            if (firstErrorField) {
                (
                    document.querySelector(
                        `[name="${firstErrorField}"]`,
                    ) as HTMLElement
                )?.focus();
            }
            return;
        }
        setStep((s) => s + 1);
    }

    const handleBack = () => {
        setStep((s) => s - 1);
    };

    const onSubmit = useCallback(async (data: ApplicationFormData) => {
        setSubmitting(true);
        const result = await submitApplication(data);
        setSubmitting(false);

        if (result.success) {
            setSubmittedData(data);
        } else {
            setToast({
                message: result.message,
                type: "error",
            });
        }
    }, []);

    const handleReset = useCallback(() => {
        setSubmittedData(null);
        methods.reset();
        setStep(0);
    }, [methods]);

    const isLoading = isValidating || submitting;

    return (
        <div className="min-h-screen bg-[#FDF6EC] flex items-center justify-center px-4 py-12">
            <div className="w-full max-w-xl bg-[#FFFCF7] rounded-2xl shadow-lg p-8 border border-[#E5DDD0]">
                <div className="flex items-center justify-center gap-3 mb-6">
                    <img
                        src={logoUrl}
                        alt="Company logo"
                        className="w-10 h-10"
                    />
                    <h1 className="text-2xl font-bold text-neutral-900">
                        Craftsmen Job Application Portal
                    </h1>
                </div>

                {submittedData ? (
                    <SubmissionSuccess
                        data={submittedData}
                        onReset={handleReset}
                    />
                ) : (
                    <>
                        <nav aria-label="Form progress" className="mb-8">
                            <div className="relative flex bg-[#EDE6DA] rounded-xl p-1">
                                <div
                                    className="pointer-events-none absolute top-1 bottom-1 bg-neutral-900 rounded-lg transition-all duration-500 ease-out"
                                    style={{
                                        left: `calc(${step} * (100% - 8px) / ${STEP_LABELS.length} + 4px)`,
                                        width: `calc((100% - 8px) / ${STEP_LABELS.length})`,
                                    }}
                                    aria-hidden="true"
                                />

                                {STEP_LABELS.map((label, i) => (
                                    <div
                                        key={label}
                                        role="listitem"
                                        aria-current={
                                            i === step ? "step" : undefined
                                        }
                                        className={`relative z-10 flex-1 py-2.5 flex items-center justify-center gap-2 text-sm transition-colors duration-500
                                    ${
                                        i === step
                                            ? "text-[#FDF6EC] font-semibold"
                                            : i < step
                                              ? "text-neutral-700 font-medium"
                                              : "text-neutral-400"
                                    }`}
                                    >
                                        {i < step ? "✓" : i + 1}
                                        <span className="hidden sm:inline">
                                            {label}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </nav>

                        <FormProvider {...methods}>
                            <form onSubmit={handleSubmit(onSubmit)}>
                                {step === 0 && <PersonalInfo />}
                                {step === 1 && <Experience />}
                                {step === 2 && <ReviewSubmit />}

                                <div className="mt-8 flex justify-between">
                                    {step > 0 ? (
                                        <button
                                            type="button"
                                            onClick={handleBack}
                                            disabled={isLoading}
                                            className="px-5 py-2 rounded-lg border border-[#D6CDBF]
                             text-sm font-medium text-neutral-700
                             hover:bg-[#F5EFE4] disabled:opacity-50
                             disabled:cursor-not-allowed"
                                        >
                                            Back
                                        </button>
                                    ) : (
                                        <div />
                                    )}

                                    {step < STEP_LABELS.length - 1 ? (
                                        <button
                                            type="button"
                                            onClick={handleNext}
                                            disabled={isLoading}
                                            aria-busy={isValidating}
                                            className="px-5 py-2 rounded-lg bg-neutral-900 text-[#FDF6EC]
                             text-sm font-semibold hover:bg-neutral-800
                             disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {isValidating
                                                ? "Checking…"
                                                : "Next →"}
                                        </button>
                                    ) : (
                                        <button
                                            type="submit"
                                            disabled={isLoading}
                                            aria-busy={submitting}
                                            className="px-5 py-2 rounded-lg bg-neutral-900 text-[#FDF6EC]
                             text-sm font-semibold hover:bg-neutral-800
                             disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {submitting
                                                ? "Submitting…"
                                                : "Submit"}
                                        </button>
                                    )}
                                </div>
                            </form>
                        </FormProvider>
                    </>
                )}
            </div>

            <Toast
                message={toast.message}
                type={toast.type}
                onClose={() => setToast({ message: "", type: null })}
            />
        </div>
    );
}
