param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

function Write-Green  { param($msg) Write-Host $msg -ForegroundColor Green }
function Write-Yellow { param($msg) Write-Host $msg -ForegroundColor Yellow }
function Write-Red    { param($msg) Write-Host $msg -ForegroundColor Red }
function Write-Cyan   { param($msg) Write-Host $msg -ForegroundColor Cyan }
function Write-Sep    { param($msg) Write-Host ""; Write-Host "=== $msg ===" -ForegroundColor Cyan }

function Invoke-Build {
    Write-Yellow "Build des images Docker (premiere fois : 3-5 minutes)..."
    docker compose -f infra/docker-compose.yml build --no-cache
    if ($LASTEXITCODE -eq 0) {
        Write-Green "Build OK. Lance : .\pulse.ps1 up"
    } else {
        Write-Red "Build echoue. Lance : .\pulse.ps1 logs"
    }
}

function Invoke-Up {
    Write-Green "Demarrage des services Pulse..."
    docker compose -f infra/docker-compose.yml up -d
    Write-Yellow "Attends 30 secondes puis lance : .\pulse.ps1 health"
}

function Invoke-Down {
    docker compose -f infra/docker-compose.yml down -v
    Write-Green "Arrete. Volumes supprimes."
}

function Invoke-Stop {
    docker compose -f infra/docker-compose.yml stop
    Write-Green "Services arretes. Volumes conserves."
}

function Invoke-Ps {
    docker compose -f infra/docker-compose.yml ps
}

function Invoke-Logs {
    docker compose -f infra/docker-compose.yml logs -f
}

function Invoke-LogsCollector {
    Write-Yellow "Logs Collector (Ctrl+C pour quitter)..."
    docker compose -f infra/docker-compose.yml logs -f collector
}

function Invoke-LogsGateway {
    Write-Yellow "Logs Gateway (Ctrl+C pour quitter)..."
    docker compose -f infra/docker-compose.yml logs -f gateway
}

function Invoke-Health {
    Write-Sep "Collector :8081"
    try {
        $r = Invoke-RestMethod -Uri 'http://localhost:8081/health' -Method GET
        $r | ConvertTo-Json -Depth 3
        if ($r.status -eq 'healthy') { Write-Green "Collector OK" }
        else { Write-Yellow "Collector DEGRADE" }
    } catch {
        Write-Red "Collector non disponible"
        Write-Yellow "Lance : .\pulse.ps1 logs-collector"
    }

    Write-Sep "Gateway :3000"
    try {
        $r = Invoke-RestMethod -Uri 'http://localhost:3000/health' -Method GET
        $r | ConvertTo-Json -Depth 3
        if ($r.status -eq 'healthy') { Write-Green "Gateway OK" }
        else { Write-Yellow "Gateway DEGRADE" }
    } catch {
        Write-Red "Gateway non disponible"
        Write-Yellow "Lance : .\pulse.ps1 logs-gateway"
    }
}

function Invoke-IngestTest {
    Write-Sep "Test ingestion de metrique"
    $headers = @{
        'Content-Type' = 'application/json'
        'X-API-Key'    = 'test-key-tenant-1'
    }
    $body = '{"name":"cpu.usage","value":87.5,"tags":{"host":"web-01","env":"prod"}}'
    try {
        $r = Invoke-RestMethod `
            -Uri     'http://localhost:3000/api/metrics/ingest' `
            -Method  POST `
            -Headers $headers `
            -Body    $body
        Write-Green "Succes !"
        $r | ConvertTo-Json
    } catch {
        Write-Red "Erreur : $($_.Exception.Message)"
        Write-Yellow "Verifie que les services tournent : .\pulse.ps1 health"
    }
}

function Invoke-WrongKeyTest {
    Write-Sep "Test mauvaise API Key - doit retourner 401"
    $headers = @{
        'Content-Type' = 'application/json'
        'X-API-Key'    = 'mauvaise-cle'
    }
    try {
        Invoke-RestMethod `
            -Uri     'http://localhost:3000/api/metrics/ingest' `
            -Method  POST `
            -Headers $headers `
            -Body    '{"name":"test","value":1}'
        Write-Red "PROBLEME : aurait du retourner 401"
    } catch {
        Write-Green "OK : 401 Unauthorized recu - la securite fonctionne"
    }
}

function Invoke-LoadTest {
    Write-Sep "Load test : 50 metriques en parallele"
    Write-Yellow "Envoi en cours..."
    $jobs = 1..50 | ForEach-Object {
        $val = $_
        Start-Job -ScriptBlock {
            param($v)
            $h = @{ 'Content-Type' = 'application/json'; 'X-API-Key' = 'test-key-tenant-1' }
            $b = '{"name":"load.test","value":' + $v + '}'
            try {
                Invoke-RestMethod -Uri 'http://localhost:3000/api/metrics/ingest' `
                    -Method POST -Headers $h -Body $b | Out-Null
                'OK'
            } catch { 'ERR' }
        } -ArgumentList $val
    }
    $results = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job
    $ok  = ($results | Where-Object { $_ -eq 'OK' }).Count
    $err = ($results | Where-Object { $_ -ne 'OK' }).Count
    Write-Green "Termine : $ok succes, $err erreurs"
}

function Invoke-DbMetrics {
    Write-Sep "5 dernieres metriques en base"
    docker exec pulse_postgres psql -U pulse -d pulse_db `
        -c "SELECT time, name, value, tags FROM metrics ORDER BY time DESC LIMIT 5;"
}

function Invoke-DbReset {
    Write-Yellow "Suppression de toutes les metriques..."
    docker exec pulse_postgres psql -U pulse -d pulse_db -c "TRUNCATE metrics;"
    Write-Green "Fait."
}

function Invoke-DbShell {
    Write-Yellow "psql interactif (quitter avec \q)..."
    docker exec -it pulse_postgres psql -U pulse -d pulse_db
}

function Invoke-Grafana {
    Write-Green "Grafana -> http://localhost:3001  login: admin / admin"
    Start-Process 'http://localhost:3001'
}

function Invoke-RabbitMQ {
    Write-Green "RabbitMQ -> http://localhost:15672  login: pulse / pulse_secret"
    Start-Process 'http://localhost:15672'
}

function Invoke-Prometheus {
    Write-Green "Prometheus -> http://localhost:9090"
    Start-Process 'http://localhost:9090'
}

function Show-Help {
    Write-Cyan "Pulse Phase 1 - PowerShell Windows"
    Write-Cyan "==================================="
    Write-Host ""
    Write-Host "DEMARRAGE (dans l'ordre)" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 build            Build les images Docker (1 seule fois)"
    Write-Host "  .\pulse.ps1 up               Demarrer les services"
    Write-Host "  .\pulse.ps1 health           Verifier que tout tourne"
    Write-Host ""
    Write-Host "QUOTIDIEN" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 up               Demarrer"
    Write-Host "  .\pulse.ps1 stop             Arreter sans perdre les donnees"
    Write-Host "  .\pulse.ps1 down             Arreter + supprimer les volumes"
    Write-Host "  .\pulse.ps1 ps               Etat des containers"
    Write-Host ""
    Write-Host "LOGS" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 logs             Tous les logs en live"
    Write-Host "  .\pulse.ps1 logs-collector   Logs du Collector Go"
    Write-Host "  .\pulse.ps1 logs-gateway     Logs du Gateway NestJS"
    Write-Host ""
    Write-Host "TESTS" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 ingest-test      Envoyer une metrique (cle valide)"
    Write-Host "  .\pulse.ps1 wrong-key-test   Tester le rejet d'une mauvaise cle"
    Write-Host "  .\pulse.ps1 load-test        50 metriques en parallele"
    Write-Host ""
    Write-Host "BASE DE DONNEES" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 db-metrics       5 dernieres metriques inserees"
    Write-Host "  .\pulse.ps1 db-shell         psql interactif"
    Write-Host "  .\pulse.ps1 db-reset         Vider la table metrics"
    Write-Host ""
    Write-Host "INTERFACES WEB" -ForegroundColor Yellow
    Write-Host "  .\pulse.ps1 grafana          :3001  admin/admin"
    Write-Host "  .\pulse.ps1 rabbitmq         :15672 pulse/pulse_secret"
    Write-Host "  .\pulse.ps1 prometheus       :9090"
    Write-Host ""
    Write-Host "CLE API DE TEST" -ForegroundColor Yellow
    Write-Host "  Header : X-API-Key: test-key-tenant-1"
}

switch ($Command) {
    'build'          { Invoke-Build }
    'up'             { Invoke-Up }
    'down'           { Invoke-Down }
    'stop'           { Invoke-Stop }
    'ps'             { Invoke-Ps }
    'logs'           { Invoke-Logs }
    'logs-collector' { Invoke-LogsCollector }
    'logs-gateway'   { Invoke-LogsGateway }
    'health'         { Invoke-Health }
    'ingest-test'    { Invoke-IngestTest }
    'wrong-key-test' { Invoke-WrongKeyTest }
    'load-test'      { Invoke-LoadTest }
    'db-metrics'     { Invoke-DbMetrics }
    'db-reset'       { Invoke-DbReset }
    'db-shell'       { Invoke-DbShell }
    'grafana'        { Invoke-Grafana }
    'rabbitmq'       { Invoke-RabbitMQ }
    'prometheus'     { Invoke-Prometheus }
    default          { Show-Help }
}
