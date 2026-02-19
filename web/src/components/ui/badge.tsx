import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground",
        outline:
          "text-foreground border-border",
        success:
          "border-transparent bg-success text-accent-foreground",
        warning:
          "border-transparent bg-warning text-white",
        muted:
          "border-transparent bg-muted text-muted-foreground",
        // Severity variants for probe findings
        critical:
          "border-transparent bg-destructive text-destructive-foreground font-semibold",
        high:
          "border-transparent bg-primary text-primary-foreground",
        medium:
          "border-transparent bg-warning text-white",
        low:
          "border-transparent bg-secondary text-secondary-foreground",
        info:
          "border-transparent bg-muted text-muted-foreground",
        // Status variants
        running:
          "border-transparent bg-secondary/20 text-secondary",
        complete:
          "border-transparent bg-success/20 text-success",
        failed:
          "border-transparent bg-destructive/20 text-destructive",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }
