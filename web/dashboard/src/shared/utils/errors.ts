export function errorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return "Something went wrong while talking to the PreDeploy Guard API.";
}
