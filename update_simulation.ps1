$jsonPath = "c:\Users\santh\programming\proj\War Room\Backend\internal\data\simulation.json"
$json = Get-Content $jsonPath -Raw | ConvertFrom-Json

$proficiencyMap = @{
    "Q_NEG2_5_A" = @{ proficiency = 3; feedback = "Excellent framing. Identifying a critical pain point signals strong problem-solution fit and high willingness to pay." }
    "Q_NEG2_5_B" = @{ proficiency = 2; feedback = "Good. Desire-based products can build loyal communities, though you must work harder to demonstrate urgency." }
    "Q_NEG2_5_C" = @{ proficiency = 2; feedback = "Honest assessment. Identify the sub-segment where it is a critical need and validate there first." }
    "Q_NEG2_5_D" = @{ proficiency = 1; feedback = "Caution. Nice-to-have products face high churn risk. Pivot to core pain points or find a market where this is a true need." }
    "Q_NEG2_7_A" = @{ proficiency = 1; feedback = "One-time sales limit growth. Consider adding a subscription layer for recurring revenue and higher LTV." }
    "Q_NEG2_7_B" = @{ proficiency = 3; feedback = "Outstanding. Subscription models deliver predictable MRR, strong LTV, and investor confidence. Excellent strategic choice." }
    "Q_NEG2_7_C" = @{ proficiency = 2; feedback = "Good for early revenue. Service models can be profitable but are hard to scale. Plan to productise over time." }
    "Q_NEG2_7_D" = @{ proficiency = 3; feedback = "High ceiling. Marketplace models create network effects, though they require solving the cold-start problem effectively." }
    "Q_NEG2_7_E" = @{ proficiency = 3; feedback = "Smart. Freemium drives mass adoption while the premium tier captures high-value users. Keep the upgrade path clear." }
    "Q_NEG2_7_F" = @{ proficiency = 1; feedback = "Honest but risky. Revenue model ambiguity is one of the top startup failure points. Commit to a hypothesis and test before Stage 1." }
}

foreach ($stage in $json.stages) {
    foreach ($q in $stage.questions) {
        if ($q.type -eq "multiple_choice") {
            foreach ($opt in $q.options) {
                if ($proficiencyMap.ContainsKey($opt.id)) {
                    $map = $proficiencyMap[$opt.id]
                    $opt | Add-Member -NotePropertyName "proficiency" -NotePropertyValue $map.proficiency -Force
                    $opt | Add-Member -NotePropertyName "feedback" -NotePropertyValue $map.feedback -Force
                } else {
                    if (-not $opt.PSObject.Properties["proficiency"]) {
                        $opt | Add-Member -NotePropertyName "proficiency" -NotePropertyValue 2 -Force
                        $opt | Add-Member -NotePropertyName "feedback" -NotePropertyValue "Reasonable choice. Your mentor will provide detailed feedback after phase evaluation." -Force
                    }
                }
            }
        }
    }
}

$t1 = [PSCustomObject]@{ from_stage="STAGE_1_VALIDATION"; to_stage="STAGE_2A_GROWTH"; scenario=[PSCustomObject]@{ title="Funded Competitor Appears"; setup="Just as your idea gained traction, a well-funded competitor enters the same market. They have a 12-month runway, a larger team, and are offering a free version to undercut your pricing. Your early users are being aggressively targeted."; leader_prompt="You have just heard the news. How do you respond to this competitive threat without losing momentum or your early adopter base?" } }
$t2 = [PSCustomObject]@{ from_stage="STAGE_2A_GROWTH"; to_stage="STAGE_2B_EXPANSION"; scenario=[PSCustomObject]@{ title="Team Burnout Signal"; setup="Your top two team members have pulled you aside. They are exhausted, feel undervalued, and one is considering leaving. Revenue is growing but the team is stretched thin. The culture from the early days is cracking under pressure."; leader_prompt="This is a pivotal moment for your startup culture. What specific steps will you take in the next 30 days to address burnout while maintaining growth momentum?" } }
$t3 = [PSCustomObject]@{ from_stage="STAGE_2B_EXPANSION"; to_stage="STAGE_3_SCALE"; scenario=[PSCustomObject]@{ title="Partnership With Strings Attached"; setup="A major strategic partner has offered you a distribution deal that could 10x your reach overnight. However, they want exclusivity for 18 months, a 30 percent revenue share, and the right to acquire you at a predetermined valuation cap. Your investors are split on this."; leader_prompt="Walk me through your decision framework. Would you take this deal, negotiate it, or walk away and why?" } }
$t4 = [PSCustomObject]@{ from_stage="STAGE_3_SCALE"; to_stage="STAGE_WARROOM_PREP"; scenario=[PSCustomObject]@{ title="Public Reputation Attack"; setup="A viral social media post amplified by a former employee claims your product caused harm and that your company lacks ethical standards. The post has 50,000 shares in 24 hours. Journalists are reaching out. Your churn has spiked 15 percent overnight."; leader_prompt="You have 2 hours before the press conference. What is your public statement strategy and your internal recovery plan for the next 72 hours?" } }

$json | Add-Member -NotePropertyName "phase_transitions" -NotePropertyValue @($t1,$t2,$t3,$t4) -Force

$json | ConvertTo-Json -Depth 20 | Set-Content -Path $jsonPath -Encoding UTF8

Write-Host "Done. File size: $((Get-Item $jsonPath).Length) bytes"
