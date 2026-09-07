# AWS Deployment Guide — Video Transcription Service

This guide provides simple, step-by-step instructions to deploy the entire distributed video transcription system on Amazon Web Services (AWS) using our pre-configured CloudFormation template.

> [!TIP]
> **100% Free Tier Eligible for New Accounts:**  
> If your AWS account is in its first 12 months, this deployment qualifies for the AWS Free Tier (750 hours/month of `t3.micro`/`t2.micro`, 5 GB S3 storage, and 1,000,000 SQS messages at **\$0.00/month**).

---

## What Gets Provisioned

Our single-click CloudFormation template (`cloudformation-ec2.yaml`) provisions:
1. **AWS S3 Bucket**: Private, secure bucket with CORS configuration for direct browser video uploads and streaming.
2. **AWS SQS Queues**: Main job queue (`transcribe-jobs`) + Dead-Letter Queue (`transcribe-jobs-dlq`) with automated 3x retry policy.
3. **IAM Instance Profile**: Grants the server least-privilege permissions to S3 and SQS. **Zero hardcoded credentials or access keys are stored on the server!**
4. **VPC & Security Group**: Isolated network exposing only port 80 (HTTP) and port 22 (SSH).
5. **EC2 Instance**: `t3.micro` (or larger) running Ubuntu 24.04 with Docker, 2GB swap space, auto-booting all 4 microservices via Docker Compose (`frontend`, `api`, `worker`, `postgres`).

---

## Step-by-Step Deployment (AWS Web Console)

### Step 1: Sign in to AWS
1. Go to [https://console.aws.amazon.com](https://console.aws.amazon.com) and log in.
2. In the top-right navigation bar, select your preferred AWS Region (for example, **US East (N. Virginia) `us-east-1`** or **Asia Pacific (Mumbai) `ap-south-1`**).

---

### Step 2: Create the CloudFormation Stack
1. In the top search bar, type `CloudFormation` and click the **CloudFormation** service.
2. Click the orange **Create stack** button (select **With new resources (standard)**).
3. Under **Prerequisite - Prepare template**, select **Template is ready**.
4. Under **Specify template**, select **Upload a template file**.
5. Click **Choose file** and browse to this file in your project:
   ```
   deploy/aws/cloudformation-ec2.yaml
   ```
6. Click **Next**.

---

### Step 3: Specify Stack Details
1. **Stack name**: Enter a name, e.g., `video-transcription`.
2. **Parameters**:
   - `InstanceType`: Keep `t3.micro` (Free Tier) or select `t3.small` / `t3.medium`.
   - `GroqApiKey`: Paste your Groq API key (`gsk_...`) so that Whisper speech-to-text transcription works immediately.
   - `GitBranch`: Keep `master`.
3. Click **Next**.
4. On the **Configure stack options** page, leave defaults and click **Next**.

---

### Step 4: Review and Deploy
1. Scroll to the bottom of the Review page.
2. Under the **Capabilities** banner, check the acknowledgment box:
   - ☑ **"I acknowledge that AWS CloudFormation might create IAM resources."**
   *(This allows CloudFormation to create the secure IAM role for S3 and SQS access).*
3. Click **Submit**.

---

### Step 5: Access Your Live Application
1. The stack status will display `CREATE_IN_PROGRESS`. It takes about **3 to 4 minutes** to provision the cloud infrastructure, install Docker, and boot the containers.
2. Once the status changes to **`CREATE_COMPLETE`**, click on the **Outputs** tab.
3. Click the link next to **`WebsiteURL`** (for example: `http://3.85.120.45`).
4. The React + HeroUI application is live! You can now:
   - Register a new account.
   - Upload videos directly to private AWS S3.
   - Watch the Go worker transcribe via Groq Whisper and stream the video with interactive synchronized subtitles.

---

## Keeping Costs at \$0.00 (or Under \$1.50 / Month)

If you only use this service occasionally (once or twice a month), follow these cost-saving tips:

### 1. Stopping the Server when Not in Use
When the EC2 instance is stopped, AWS charges **\$0.00 for compute (CPU & RAM)**:
1. In the AWS Console, search for **EC2** → click **Instances**.
2. Select your instance (`transcribe-server`).
3. Click **Instance state** → **Stop instance**.
4. While stopped, you only pay for the 25 GB EBS storage (~**\$1.60/month**, or **\$0.00** if your account is in the 12-month free tier).

### 2. Starting the Server Again
When you want to demo or use the application:
1. Go to EC2 → Instances → Select `transcribe-server`.
2. Click **Instance state** → **Start instance**.
3. In ~30 seconds, the server boots up. The systemd service automatically starts all 4 Docker containers, and your application is live again at its public IP!

### 3. Deleting Everything (Complete Cleanup)
If you no longer need the project on AWS:
1. Go to **CloudFormation** → Select `video-transcription`.
2. Click **Delete**.
3. CloudFormation will delete the EC2 instance, IAM roles, Security Groups, SQS queues, and VPC in one click.
*(Note: To delete the S3 bucket if it has videos inside, empty the bucket in the S3 console first, then delete the stack).*
