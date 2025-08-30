# Deploy Fabric API to Railway

This guide will help you deploy the Fabric AI framework as a web-accessible API on Railway.

## Prerequisites

1. A [Railway](https://railway.app) account
2. AI provider API keys (OpenAI, Anthropic, Gemini, etc.)
3. Git repository with your Fabric code

## Quick Deploy

### Option 1: Deploy from GitHub (Recommended)

1. **Fork or clone this repository** to your GitHub account

2. **Create a new Railway project**
   - Go to [Railway](https://railway.app)
   - Click "New Project" → "Deploy from GitHub repo"
   - Select your Fabric repository
   - Railway will automatically detect the Dockerfile and deploy

3. **Set environment variables** in Railway:
   ```
   OPENAI_API_KEY=your_openai_api_key_here
   ANTHROPIC_API_KEY=your_anthropic_api_key_here
   GEMINI_API_KEY=your_gemini_api_key_here
   FABRIC_API_KEY=your_secure_api_key_for_endpoints (optional but recommended)
   ```

4. **Deploy** - Railway will automatically build and deploy your API

### Option 2: Deploy with Railway CLI

1. **Install Railway CLI**
   ```bash
   npm install -g @railway/cli
   # or
   curl -fsSL https://railway.app/install.sh | sh
   ```

2. **Login to Railway**
   ```bash
   railway login
   ```

3. **Deploy from your local repository**
   ```bash
   cd /path/to/fabric
   railway init
   railway up
   ```

4. **Set environment variables**
   ```bash
   railway variables set OPENAI_API_KEY=your_key_here
   railway variables set ANTHROPIC_API_KEY=your_key_here
   railway variables set GEMINI_API_KEY=your_key_here
   railway variables set FABRIC_API_KEY=your_secure_api_key
   ```

## Environment Variables

### Required AI Provider Keys (at least one)
- `OPENAI_API_KEY` - For GPT models
- `ANTHROPIC_API_KEY` - For Claude models  
- `GEMINI_API_KEY` - For Gemini models
- `GROQ_API_KEY` - For Groq models
- `COHERE_API_KEY` - For Cohere models
- `PERPLEXITY_API_KEY` - For Perplexity models

### Optional Security
- `FABRIC_API_KEY` - Secures all API endpoints (highly recommended for production)

### Optional Configuration
- `DEFAULT_MODEL` - Default model to use (e.g., "gpt-4o-mini")
- `DEFAULT_VENDOR` - Default vendor (e.g., "openai")

## API Endpoints

Once deployed, your API will be available at `https://your-railway-app.railway.app`

### Key Endpoints

**Pattern Management:**
- `GET /patterns/names` - List all available patterns
- `GET /patterns/:name` - Get specific pattern
- `POST /patterns/:name/apply` - Apply pattern to input

**AI Chat:**
- `POST /chat` - Chat with AI using patterns

**Configuration:**
- `GET /models/names` - List available models
- `GET /config` - Get current configuration

**YouTube Processing:**
- `POST /youtube/transcript` - Extract YouTube transcripts

### Example API Usage

1. **List available patterns:**
   ```bash
   curl https://your-app.railway.app/patterns/names
   ```

2. **Chat with AI using a pattern:**
   ```bash
   curl -X POST https://your-app.railway.app/chat \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_FABRIC_API_KEY" \
     -d '{
       "prompts": [{
         "userInput": "Explain quantum computing",
         "patternName": "explain_terms",
         "model": "gpt-4o-mini",
         "vendor": "openai"
       }]
     }'
   ```

3. **Get available models:**
   ```bash
   curl https://your-app.railway.app/models/names
   ```

## Security Considerations

1. **Always set `FABRIC_API_KEY`** for production deployments
2. **Keep your AI provider keys secure** - never expose them in client-side code
3. **Monitor usage** to avoid unexpected API costs
4. **Consider rate limiting** for public APIs

## Troubleshooting

### Common Issues

1. **Build fails**: Check that all dependencies in `go.mod` are accessible
2. **API returns 500**: Check Railway logs for missing environment variables
3. **AI requests fail**: Verify your AI provider API keys are correct
4. **403 Forbidden**: Make sure you're including the `Authorization: Bearer YOUR_FABRIC_API_KEY` header if set

### View Logs
```bash
railway logs
```

### Connect to deployed service
```bash
railway shell
```

## Cost Optimization

1. **Use Railway's free tier** for development/testing
2. **Monitor AI provider usage** - costs can add up quickly
3. **Set up billing alerts** in your AI provider dashboards
4. **Consider using cheaper models** like GPT-4o-mini for development

## Advanced Configuration

### Custom Domain
1. Go to your Railway project settings
2. Add your custom domain
3. Configure DNS as instructed

### Scaling
Railway automatically handles scaling, but you can:
1. Monitor resource usage in the Railway dashboard
2. Upgrade your plan if needed for higher limits

## File Structure

The deployment uses these key files:
- `Dockerfile` - Container configuration
- `railway.toml` - Railway deployment settings
- `.env.example` - Environment variable template

## Support

- [Railway Documentation](https://docs.railway.app/)
- [Fabric GitHub Issues](https://github.com/danielmiessler/fabric/issues)
- [Railway Discord](https://discord.gg/railway)