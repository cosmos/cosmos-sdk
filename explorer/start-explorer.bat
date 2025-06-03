@echo off
echo 🌌 Starting Tajeor Blockchain Explorer...
echo ========================================

echo.
echo 📦 Installing dependencies...
call npm install

echo.
echo 🚀 Starting the explorer API server...
echo 📊 Explorer will be available at: http://localhost:3000
echo 🔗 API endpoints: http://localhost:3000/api/
echo.
echo ⚠️  Make sure your Tajeor blockchain is initialized and has validators!
echo 💡 Press Ctrl+C to stop the server
echo.

call npm start 