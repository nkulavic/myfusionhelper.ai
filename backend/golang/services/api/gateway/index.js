exports.handler = async (event) => {
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      success: true,
      message: "Service is healthy",
      data: {
        service: "mfh-api-gateway",
        version: "1.0.0",
        stage: process.env.STAGE || "dev",
        timestamp: new Date().toISOString(),
      },
    }),
  };
};
