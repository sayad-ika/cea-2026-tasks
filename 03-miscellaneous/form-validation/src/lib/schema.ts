import { z } from "zod";

const jobSchema = z
    .object({
        company: z.string().min(1, "Company name is required"),
        role: z.string().min(1, "Role is required"),
        startDate: z.string().min(1, "Start date is required"),
        endDate: z.string().optional().or(z.literal("")),
        isCurrent: z.boolean(),
    })
    .refine(
        (data) => {
            if (data.isCurrent) return true;
            return !!data.endDate && data.endDate !== "";
        },
        { message: "End date is required", path: ["endDate"] },
    )
    .refine(
        (data) => {
            if (data.isCurrent) return true;
            if (!data.endDate) return true;
            return new Date(data.endDate) > new Date(data.startDate);
        },
        { message: "End Date must be after start date", path: ["endDate"] },
    );

export const applicationSchema = z.object({
    firstName: z.string().min(1, "First name is required"),
    lastName: z.string().min(1, "Last name is required"),
    email: z.email("Must be a valid email address").min(1, "Email is required"),
    phone: z
        .string()
        .min(1, "Phone number is required")
        .regex(/^\+?[0-9\s\-()]{7,15}$/, "Enter a valid phone number"),

    linkedinUrl: z.url("Enter a valid URL").optional().or(z.literal("")),
    githubUrl: z.url("Enter a valid URL").optional().or(z.literal("")),
    portfolioUrl: z.url("Enter a valid URL").optional().or(z.literal("")),

    hasExperience: z.boolean(),

    yearsOfExperience: z
        .number({ error: "Must be a number" })
        .min(0, "Cannot be negative")
        .max(50, "Value seems too high"),
    previousJobs: z.array(jobSchema),

    resume: z
        .any()
        .refine((files) => files && files.length > 0, "Resume is required")
        .refine((files) => {
            if (!files || files.length === 0) return true;
            const name = files[0]?.name ?? "";
            return [".pdf", ".docx", ".doc"].some((ext) =>
                name.toLowerCase().endsWith(ext),
            );
        }, "Only PDF, DOC, or DOCX files are allowed"),
});

export type ApplicationFormData = z.infer<typeof applicationSchema>;
