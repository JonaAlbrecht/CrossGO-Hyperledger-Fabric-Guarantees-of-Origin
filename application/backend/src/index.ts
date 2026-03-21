// Express REST API entry point for GO Platform
import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';
import { logger } from './middleware/logger';
import { errorHandler } from './middleware/error';
import authRoutes from './routes/auth';
import deviceRoutes from './routes/devices';
import guaranteeRoutes from './routes/guarantees';
import transferRoutes from './routes/transfers';
import conversionRoutes from './routes/conversions';
import cancellationRoutes from './routes/cancellations';
import queryRoutes from './routes/queries';

dotenv.config();

const app = express();
const PORT = process.env.PORT ?? 3001;

app.use(cors());
app.use(express.json());

// Request logging
app.use((req, _res, next) => {
    logger.info(`${req.method} ${req.path}`);
    next();
});

// Health check
app.get('/api/health', (_req, res) => {
    res.json({ status: 'ok', service: 'go-platform-backend' });
});

// Routes
app.use('/api/auth', authRoutes);
app.use('/api/devices', deviceRoutes);
app.use('/api/guarantees', guaranteeRoutes);
app.use('/api/transfers', transferRoutes);
app.use('/api/conversions', conversionRoutes);
app.use('/api/cancellations', cancellationRoutes);
app.use('/api/queries', queryRoutes);

// Global error handler
app.use(errorHandler);

app.listen(PORT, () => {
    logger.info(`GO Platform backend listening on port ${PORT}`);
});

export default app;
