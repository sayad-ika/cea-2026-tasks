import { useFormContext } from "react-hook-form";
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

export function ReviewSubmit() {
    const { getValues } = useFormContext<ApplicationFormData>();
    const values = getValues();
    const resumeFile = values.resume?.[0] as File | undefined;

    return (
        <div className="space-y-6">
            <h2 className="text-xl font-semibold text-neutral-900">
                Review Your Application
            </h2>

            <section aria-labelledby="review-personal">
                <h3
                    id="review-personal"
                    className="text-base font-semibold text-neutral-600 mb-2"
                >
                    Personal Info
                </h3>
                <div className="rounded-lg border border-[#E5DDD0] px-4 py-1">
                    <Row label="First Name" value={values.firstName} />
                    <Row label="Last Name" value={values.lastName} />
                    <Row label="Email" value={values.email} />
                    <Row label="Phone" value={values.phone ?? ""} />
                    {values.linkedinUrl && (
                        <Row label="LinkedIn" value={values.linkedinUrl} />
                    )}
                    {values.githubUrl && (
                        <Row label="GitHub" value={values.githubUrl} />
                    )}
                    {values.portfolioUrl && (
                        <Row label="Portfolio" value={values.portfolioUrl} />
                    )}
                </div>
            </section>

            <section aria-labelledby="review-experience">
                <h3
                    id="review-experience"
                    className="text-base font-semibold text-neutral-600 mb-2"
                >
                    Experience
                </h3>
                <div className="rounded-lg border border-[#E5DDD0] px-4 py-1">
                    <Row label="Resume" value={resumeFile?.name ?? ""} />
                </div>

                {values.hasExperience ? (
                    <>
                        <div className="mt-3 rounded-lg border border-[#E5DDD0] px-4 py-1">
                            <Row
                                label="Years of Experience"
                                value={values.yearsOfExperience}
                            />
                        </div>

                        {values.previousJobs?.length > 0 && (
                            <div className="mt-3 space-y-3">
                                {values.previousJobs.map((job, i) => (
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
                                        <Row label="Role" value={job.role} />
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
    );
}
