import type { ApplicationFormData } from "../lib/schema";

function Row({ label, value }: { label: string; value: string | number }) {
    return (
        <div className="flex justify-between py-2 border-b border-[#EDE6DA] text-sm">
            <span className="font-medium text-neutral-500">{label}</span>
            <span className="text-neutral-900">
                {value !== "" ? value : "Not provided"}
            </span>
        </div>
    );
}

interface SubmissionSuccessProps {
    data: ApplicationFormData;
    onReset: () => void;
}

export function SubmissionSuccess({ data, onReset }: SubmissionSuccessProps) {
    const resumeFile = data.resume?.[0] as File | undefined;

    return (
        <div>
            <div className="text-center mb-8">
                <div className="mx-auto w-16 h-16 rounded-full bg-neutral-900 flex items-center justify-center mb-4">
                    <svg
                        className="w-8 h-8 text-[#FDF6EC]"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        strokeWidth={2.5}
                    >
                        <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            d="M5 13l4 4L19 7"
                        />
                    </svg>
                </div>
                <h2 className="text-2xl font-bold text-neutral-900 mb-2">
                    Application Submitted!
                </h2>
                <p className="text-neutral-500">
                    Thank you for your submission. Here is a summary of your
                    application.
                </p>
            </div>

            <div className="space-y-6">
                <section aria-labelledby="success-personal">
                    <h3
                        id="success-personal"
                        className="text-base font-semibold text-neutral-600 mb-2"
                    >
                        Personal Info
                    </h3>
                    <div className="rounded-lg border border-[#E5DDD0] px-4 py-1">
                        <Row label="First Name" value={data.firstName} />
                        <Row label="Last Name" value={data.lastName} />
                        <Row label="Email" value={data.email} />
                        <Row label="Phone" value={data.phone ?? ""} />
                        {data.linkedinUrl && (
                            <Row label="LinkedIn" value={data.linkedinUrl} />
                        )}
                        {data.githubUrl && (
                            <Row label="GitHub" value={data.githubUrl} />
                        )}
                        {data.portfolioUrl && (
                            <Row label="Portfolio" value={data.portfolioUrl} />
                        )}
                    </div>
                </section>

                <section aria-labelledby="success-experience">
                    <h3
                        id="success-experience"
                        className="text-base font-semibold text-neutral-600 mb-2"
                    >
                        Experience
                    </h3>
                    <div className="rounded-lg border border-[#E5DDD0] px-4 py-1">
                        <Row label="Resume" value={resumeFile?.name ?? ""} />
                    </div>

                    {data.hasExperience ? (
                        <>
                            <div className="mt-3 rounded-lg border border-[#E5DDD0] px-4 py-1">
                                <Row
                                    label="Years of Experience"
                                    value={data.yearsOfExperience}
                                />
                            </div>

                            {data.previousJobs?.length > 0 && (
                                <div className="mt-3 space-y-3">
                                    {data.previousJobs.map((job, i) => (
                                        <div
                                            key={i}
                                            className="rounded-lg border border-[#E5DDD0] px-4 py-1 bg-[#F5EFE4]"
                                        >
                                            <p className="text-xs font-bold text-neutral-500 uppercase pt-2">
                                                Job #{i + 1}
                                            </p>
                                            <Row
                                                label="Company"
                                                value={job.company}
                                            />
                                            <Row
                                                label="Role"
                                                value={job.role}
                                            />
                                            <Row
                                                label="Start Date"
                                                value={job.startDate}
                                            />
                                            <Row
                                                label="End Date"
                                                value={
                                                    job.isCurrent
                                                        ? "Present"
                                                        : (job.endDate ?? "")
                                                }
                                            />
                                        </div>
                                    ))}
                                </div>
                            )}
                        </>
                    ) : (
                        <div className="mt-3 rounded-lg border border-[#E5DDD0] px-4 py-3">
                            <p className="text-sm text-neutral-500 italic">
                                No prior experience
                            </p>
                        </div>
                    )}
                </section>
            </div>

            <div className="mt-8 text-center">
                <button
                    type="button"
                    onClick={onReset}
                    className="px-6 py-2.5 rounded-lg bg-neutral-900 text-[#FDF6EC]
                     text-sm font-semibold hover:bg-neutral-800"
                >
                    Submit Another Application
                </button>
            </div>
        </div>
    );
}
