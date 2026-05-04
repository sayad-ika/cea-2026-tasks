import { useFormContext, useFieldArray } from "react-hook-form";
import type { ApplicationFormData } from "../lib/schema";
import { FieldError } from "./FieldError";

export function Experience() {
    const {
        register,
        control,
        setValue,
        clearErrors,
        watch,
        formState: { errors },
    } = useFormContext<ApplicationFormData>();

    const hasExperience = watch("hasExperience");

    const { fields, append, remove, replace } = useFieldArray({
        control,
        name: "previousJobs",
    });

    return (
        <fieldset className="space-y-6">
            <legend className="text-xl font-semibold text-neutral-900 mb-4">
                Experience
            </legend>

            <div>
                <label className="flex items-center gap-2 cursor-pointer">
                    <input
                        type="checkbox"
                        {...register("hasExperience", {
                            onChange: (e) => {
                                if (!e.target.checked) {
                                    setValue("yearsOfExperience", 0);
                                    replace([]);
                                }
                            },
                        })}
                        className="w-4 h-4 accent-neutral-900"
                    />
                    <span className="text-sm font-medium text-neutral-700">
                        I have prior experience
                    </span>
                </label>
            </div>

            <div>
                <label
                    htmlFor="resume"
                    className="block text-sm font-medium text-neutral-700"
                >
                    Resume{" "}
                    <span aria-hidden="true" className="text-red-500">
                        *
                    </span>
                </label>
                <input
                    id="resume"
                    type="file"
                    accept=".pdf,.docx,.doc"
                    {...register("resume")}
                    className={`mt-1 block w-full text-sm
            file:mr-4 file:py-2 file:px-4
            file:rounded-md file:border-0
            file:text-sm file:font-semibold
            file:bg-neutral-900 file:text-[#FDF6EC]
            hover:file:bg-neutral-800 file:cursor-pointer
            ${errors.resume ? "border border-red-500 rounded-md" : ""}`}
                />
                <p className="mt-1 text-xs text-neutral-500">
                    Accepted formats: PDF, DOC, DOCX
                </p>
                <FieldError
                    id="resume-error"
                    message={errors.resume?.message as string | undefined}
                />
            </div>

            {hasExperience && (
                <>
                    <div>
                        <label
                            htmlFor="yearsOfExperience"
                            className="block text-sm font-medium text-neutral-700"
                        >
                            Years of Experience{" "}
                            <span aria-hidden="true" className="text-red-500">
                                *
                            </span>
                        </label>
                        <input
                            id="yearsOfExperience"
                            type="number"
                            min={0}
                            {...register("yearsOfExperience", {
                                valueAsNumber: true,
                            })}
                            aria-invalid={!!errors.yearsOfExperience}
                            aria-describedby="yearsOfExperience-error"
                            className={`mt-1 block w-32 rounded-md border px-3 py-2 text-sm
                shadow-sm focus:outline-none focus:ring-2
                ${
                    errors.yearsOfExperience
                        ? "border-red-500 focus:ring-red-300"
                        : "border-[#D6CDBF] focus:ring-neutral-400"
                }`}
                        />
                        <FieldError
                            id="yearsOfExperience-error"
                            message={errors.yearsOfExperience?.message}
                        />
                    </div>

                    <div>
                        <h3 className="text-base font-medium text-neutral-700 mb-3">
                            Previous Jobs
                        </h3>

                        {fields.length === 0 && (
                            <p className="text-sm text-neutral-500 italic mb-3">
                                No previous jobs added yet.
                            </p>
                        )}

                        <div className="space-y-4">
                            {fields.map((field, index) => {
                                const rowErrors = errors.previousJobs?.[index];
                                const isCurrent = watch(
                                    `previousJobs.${index}.isCurrent`,
                                );

                                return (
                                    <div
                                        key={field.id}
                                        className="rounded-lg border border-[#E5DDD0] p-4 bg-[#F5EFE4]
                                   relative grid grid-cols-2 gap-3"
                                    >
                                        <p className="col-span-2 text-xs font-semibold text-neutral-500 uppercase">
                                            Job #{index + 1}
                                        </p>

                                        <div>
                                            <label
                                                htmlFor={`previousJobs.${index}.company`}
                                                className="block text-sm font-medium text-neutral-700"
                                            >
                                                Company
                                            </label>
                                            <input
                                                id={`previousJobs.${index}.company`}
                                                {...register(
                                                    `previousJobs.${index}.company`,
                                                )}
                                                aria-invalid={
                                                    !!rowErrors?.company
                                                }
                                                aria-describedby={`job-${index}-company-error`}
                                                className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                          shadow-sm focus:outline-none focus:ring-2
                          ${
                              rowErrors?.company
                                  ? "border-red-500 focus:ring-red-300"
                                  : "border-[#D6CDBF] focus:ring-neutral-400"
                          }`}
                                            />
                                            <FieldError
                                                id={`job-${index}-company-error`}
                                                message={
                                                    rowErrors?.company?.message
                                                }
                                            />
                                        </div>

                                        <div>
                                            <label
                                                htmlFor={`previousJobs.${index}.role`}
                                                className="block text-sm font-medium text-neutral-700"
                                            >
                                                Role
                                            </label>
                                            <input
                                                id={`previousJobs.${index}.role`}
                                                {...register(
                                                    `previousJobs.${index}.role`,
                                                )}
                                                aria-invalid={!!rowErrors?.role}
                                                aria-describedby={`job-${index}-role-error`}
                                                className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                          shadow-sm focus:outline-none focus:ring-2
                          ${
                              rowErrors?.role
                                  ? "border-red-500 focus:ring-red-300"
                                  : "border-[#D6CDBF] focus:ring-neutral-400"
                          }`}
                                            />
                                            <FieldError
                                                id={`job-${index}-role-error`}
                                                message={
                                                    rowErrors?.role?.message
                                                }
                                            />
                                        </div>

                                        <div>
                                            <label
                                                htmlFor={`previousJobs.${index}.startDate`}
                                                className="block text-sm font-medium text-neutral-700"
                                            >
                                                Start Date
                                            </label>
                                            <input
                                                id={`previousJobs.${index}.startDate`}
                                                type="date"
                                                {...register(
                                                    `previousJobs.${index}.startDate`,
                                                )}
                                                aria-invalid={
                                                    !!rowErrors?.startDate
                                                }
                                                aria-describedby={`job-${index}-startDate-error`}
                                                className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                          shadow-sm focus:outline-none focus:ring-2
                          ${
                              rowErrors?.startDate
                                  ? "border-red-500 focus:ring-red-300"
                                  : "border-[#D6CDBF] focus:ring-neutral-400"
                          }`}
                                            />
                                            <FieldError
                                                id={`job-${index}-startDate-error`}
                                                message={
                                                    rowErrors?.startDate
                                                        ?.message
                                                }
                                            />
                                        </div>

                                        {isCurrent ? (
                                            <div className="flex flex-col justify-end">
                                                <span className="block text-sm font-medium text-neutral-700">
                                                    End Date
                                                </span>
                                                <span className="mt-1 px-3 py-2 text-sm text-neutral-500 italic">
                                                    Present
                                                </span>
                                            </div>
                                        ) : (
                                            <div>
                                                <label
                                                    htmlFor={`previousJobs.${index}.endDate`}
                                                    className="block text-sm font-medium text-neutral-700"
                                                >
                                                    End Date
                                                </label>
                                                <input
                                                    id={`previousJobs.${index}.endDate`}
                                                    type="date"
                                                    {...register(
                                                        `previousJobs.${index}.endDate`,
                                                    )}
                                                    aria-invalid={
                                                        !!rowErrors?.endDate
                                                    }
                                                    aria-describedby={`job-${index}-endDate-error`}
                                                    className={`mt-1 block w-full rounded-md border px-3 py-2 text-sm
                              shadow-sm focus:outline-none focus:ring-2
                              ${
                                  rowErrors?.endDate
                                      ? "border-red-500 focus:ring-red-300"
                                      : "border-[#D6CDBF] focus:ring-neutral-400"
                              }`}
                                                />
                                                <FieldError
                                                    id={`job-${index}-endDate-error`}
                                                    message={
                                                        rowErrors?.endDate
                                                            ?.message
                                                    }
                                                />
                                            </div>
                                        )}

                                        <div className="col-span-2 flex items-center justify-between">
                                            <label className="flex items-center gap-2 cursor-pointer">
                                                <input
                                                    type="checkbox"
                                                    {...register(
                                                        `previousJobs.${index}.isCurrent`,
                                                        {
                                                            onChange: (
                                                                e: React.ChangeEvent<HTMLInputElement>,
                                                            ) => {
                                                                if (
                                                                    e.target
                                                                        .checked
                                                                ) {
                                                                    setValue(
                                                                        `previousJobs.${index}.endDate`,
                                                                        "",
                                                                    );
                                                                    clearErrors(
                                                                        `previousJobs.${index}.endDate`,
                                                                    );
                                                                }
                                                            },
                                                        },
                                                    )}
                                                    className="w-4 h-4 accent-neutral-900"
                                                />
                                                <span className="text-sm text-neutral-700">
                                                    I currently work here
                                                </span>
                                            </label>

                                            <button
                                                type="button"
                                                onClick={() => remove(index)}
                                                aria-label={`Remove job #${index + 1}`}
                                                className="text-sm text-red-500 hover:text-red-700
                                         underline w-fit"
                                            >
                                                Remove
                                            </button>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>

                        <button
                            type="button"
                            onClick={() =>
                                append({
                                    company: "",
                                    role: "",
                                    startDate: "",
                                    endDate: "",
                                    isCurrent: false,
                                })
                            }
                            className="mt-3 text-sm font-medium text-neutral-900 hover:text-neutral-700
                         underline"
                        >
                            + Add previous job
                        </button>
                    </div>
                </>
            )}
        </fieldset>
    );
}
