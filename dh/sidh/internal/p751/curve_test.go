package internal

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// Sage script for generating test vectors:
// sage: p = 2^372 * 3^239 - 1; Fp = GF(p)
// sage: R.<x> = Fp[]
// sage: Fp2 = Fp.extension(x^2 + 1, 'i')
// sage: i = Fp2.gen()
// sage: A = 4385300808024233870220415655826946795549183378139271271040522089756750951667981765872679172832050962894122367066234419550072004266298327417513857609747116903999863022476533671840646615759860564818837299058134292387429068536219*i + 1408083354499944307008104531475821995920666351413327060806684084512082259107262519686546161682384352696826343970108773343853651664489352092568012759783386151707999371397181344707721407830640876552312524779901115054295865393760
// sage: C = 933177602672972392833143808100058748100491911694554386487433154761658932801917030685312352302083870852688835968069519091048283111836766101703759957146191882367397129269726925521881467635358356591977198680477382414690421049768*i + 9088894745865170214288643088620446862479558967886622582768682946704447519087179261631044546285104919696820250567182021319063155067584445633834024992188567423889559216759336548208016316396859149888322907914724065641454773776307
// sage: E = EllipticCurve(Fp2, [0,A/C,0,1,0])
// sage: XP, YP, ZP = (8172151271761071554796221948801462094972242987811852753144865524899433583596839357223411088919388342364651632180452081960511516040935428737829624206426287774255114241789158000915683252363913079335550843837650671094705509470594*i + 9326574858039944121604015439381720195556183422719505497448541073272720545047742235526963773359004021838961919129020087515274115525812121436661025030481584576474033630899768377131534320053412545346268645085054880212827284581557, 2381174772709336084066332457520782192315178511983342038392622832616744048226360647551642232950959910067260611740876401494529727990031260499974773548012283808741733925525689114517493995359390158666069816204787133942283380884077*i + 5378956232034228335189697969144556552783858755832284194802470922976054645696324118966333158267442767138528227968841257817537239745277092206433048875637709652271370008564179304718555812947398374153513738054572355903547642836171, 1)
// sage: XQ, YQ, ZQ = (58415083458086949460774631288254059520198050925769753726965257584725433404854146588020043440874302516770401547098747946831855304325931998303167820332782040123488940125615738529117852871565495005635582483848203360379709783913*i + 5253691516070829381103946367549411712646323371150903710970418364086793242952116417060222617507143384179331507713214357463046698181663711723374230655213762988214127964685312538735038552802853885461137159568726585741937776693865 : 5158855522131677856751020091923296498106687563952623634063620100943451774747711264382913073742659403103540173656392945587315798626256921403472228775673617431559609635074280718249654789362069112718012186738875122436282913402060*i + 6207982879267706771450280080677554401823496760317291100742934673980817977221212113336809046153999230028113719924529054308271353271269929392876016936836430864080079684360816968551808465413552611322477628209203066093559492022329 : 1)
// sage: P = E((XP,YP,ZP))
// sage: X2, Y2, Z2 = 2*P
// sage: X3, Y3, Z3 = 3*P
// sage: m = 96550223052359874398280314003345143371473380422728857598463622014420884224892

// A = 4385300808024233870220415655826946795549183378139271271040522089756750951667981765872679172832050962894122367066234419550072004266298327417513857609747116903999863022476533671840646615759860564818837299058134292387429068536219*i + 1408083354499944307008104531475821995920666351413327060806684084512082259107262519686546161682384352696826343970108773343853651664489352092568012759783386151707999371397181344707721407830640876552312524779901115054295865393760
var curve_A = ExtensionFieldElement{
	A: Fp751Element{0x8319eb18ca2c435e, 0x3a93beae72cd0267, 0x5e465e1f72fd5a84, 0x8617fa4150aa7272, 0x887da24799d62a13, 0xb079b31b3c7667fe, 0xc4661b150fa14f2e, 0xd4d2b2967bc6efd6, 0x854215a8b7239003, 0x61c5302ccba656c2, 0xf93194a27d6f97a2, 0x1ed9532bca75},
	B: Fp751Element{0xb6f541040e8c7db6, 0x99403e7365342e15, 0x457e9cee7c29cced, 0x8ece72dc073b1d67, 0x6e73cef17ad28d28, 0x7aed836ca317472, 0x89e1de9454263b54, 0x745329277aa0071b, 0xf623dfc73bc86b9b, 0xb8e3c1d8a9245882, 0x6ad0b3d317770bec, 0x5b406e8d502b}}

// C = 933177602672972392833143808100058748100491911694554386487433154761658932801917030685312352302083870852688835968069519091048283111836766101703759957146191882367397129269726925521881467635358356591977198680477382414690421049768*i + 9088894745865170214288643088620446862479558967886622582768682946704447519087179261631044546285104919696820250567182021319063155067584445633834024992188567423889559216759336548208016316396859149888322907914724065641454773776307
var curve_C = ExtensionFieldElement{
	A: Fp751Element{0x4fb2358bbf723107, 0x3a791521ac79e240, 0x283e24ef7c4c922f, 0xc89baa1205e33cc, 0x3031be81cff6fee1, 0xaf7a494a2f6a95c4, 0x248d251eaac83a1d, 0xc122fca1e2550c88, 0xbc0451b11b6cfd3d, 0x9c0a114ab046222c, 0x43b957b32f21f6ea, 0x5b9c87fa61de},
	B: Fp751Element{0xacf142afaac15ec6, 0xfd1322a504a071d5, 0x56bb205e10f6c5c6, 0xe204d2849a97b9bd, 0x40b0122202fe7f2e, 0xecf72c6fafacf2cb, 0x45dfc681f869f60a, 0x11814c9aff4af66c, 0x9278b0c4eea54fe7, 0x9a633d5baf7f2e2e, 0x69a329e6f1a05112, 0x1d874ace23e4}}

var curve = ProjectiveCurveParameters{A: curve_A, C: curve_C}

// x(P) = 8172151271761071554796221948801462094972242987811852753144865524899433583596839357223411088919388342364651632180452081960511516040935428737829624206426287774255114241789158000915683252363913079335550843837650671094705509470594*i + 9326574858039944121604015439381720195556183422719505497448541073272720545047742235526963773359004021838961919129020087515274115525812121436661025030481584576474033630899768377131534320053412545346268645085054880212827284581557
var affine_xP = ExtensionFieldElement{
	A: Fp751Element{0xe8d05f30aac47247, 0x576ec00c55441de7, 0xbf1a8ec5fe558518, 0xd77cb17f77515881, 0x8e9852837ee73ec4, 0x8159634ad4f44a6b, 0x2e4eb5533a798c5, 0x9be8c4354d5bc849, 0xf47dc61806496b84, 0x25d0e130295120e0, 0xdbef54095f8139e3, 0x5a724f20862c},
	B: Fp751Element{0x3ca30d7623602e30, 0xfb281eddf45f07b7, 0xd2bf62d5901a45bc, 0xc67c9baf86306dd2, 0x4e2bd93093f538ca, 0xcfd92075c25b9cbe, 0xceafe9a3095bcbab, 0x7d928ad380c85414, 0x37c5f38b2afdc095, 0x75325899a7b779f4, 0xf130568249f20fdd, 0x178f264767d1}}

// x([2]P) = 1476586462090705633631615225226507185986710728845281579274759750260315746890216330325246185232948298241128541272709769576682305216876843626191069809810990267291824247158062860010264352034514805065784938198193493333201179504845*i + 3623708673253635214546781153561465284135688791018117615357700171724097420944592557655719832228709144190233454198555848137097153934561706150196041331832421059972652530564323645509890008896574678228045006354394485640545367112224
var affine_xP2 = ExtensionFieldElement{
	A: Fp751Element{0x2a77afa8576ce979, 0xab1360e69b0aeba0, 0xd79e3e3cbffad660, 0x5fd0175aa10f106b, 0x1800ebafce9fbdbc, 0x228fc9142bdd6166, 0x867cf907314e34c3, 0xa58d18c94c13c31c, 0x699a5bc78b11499f, 0xa29fc29a01f7ccf1, 0x6c69c0c5347eebce, 0x38ecee0cc57},
	B: Fp751Element{0x43607fd5f4837da0, 0x560bad4ce27f8f4a, 0x2164927f8495b4dd, 0x621103fdb831a997, 0xad740c4eea7db2db, 0x2cde0442205096cd, 0x2af51a70ede8324e, 0x41a4e680b9f3466, 0x5481f74660b8f476, 0xfcb2f3e656ff4d18, 0x42e3ce0837171acc, 0x44238c30530c}}

// x([3]P) = 9351941061182433396254169746041546943662317734130813745868897924918150043217746763025923323891372857734564353401396667570940585840576256269386471444236630417779544535291208627646172485976486155620044292287052393847140181703665*i + 9010417309438761934687053906541862978676948345305618417255296028956221117900864204687119686555681136336037659036201780543527957809743092793196559099050594959988453765829339642265399496041485088089691808244290286521100323250273
var affine_xP3 = ExtensionFieldElement{
	A: Fp751Element{0x2096e3f23feca947, 0xf36f635aa4ad8634, 0xdae3b1c6983c5e9a, 0xe08df6c262cb74b4, 0xd2ca4edc37452d3d, 0xfb5f3fe42f500c79, 0x73740aa3abc2b21f, 0xd535fd869f914cca, 0x4a558466823fb67f, 0x3e50a7a0e3bfc715, 0xf43c6da9183a132f, 0x61aca1e1b8b9},
	B: Fp751Element{0x1e54ec26ea5077bd, 0x61380572d8769f9a, 0xc615170684f59818, 0x6309c3b93e84ef6e, 0x33c74b1318c3fcd0, 0xfe8d7956835afb14, 0x2d5a7b55423c1ecc, 0x869db67edfafea68, 0x1292632394f0a628, 0x10bba48225bfd141, 0x6466c28b408daba, 0x63cacfdb7c43}}

// x([2^2]P) = 441719501189485559222919502512761433931671682884872259563221427434901842337947564993718830905758163254463901652874331063768876314142359813382575876106725244985607032091781306919778265250690045578695338669105227100119314831452*i + 6961734028200975729170216310486458180126343885294922940439352055937945948015840788921225114530454649744697857047401608073256634790353321931728699534700109268264491160589480994022419317695690866764726967221310990488404411684053
var affine_xP4 = ExtensionFieldElement{
	A: Fp751Element{0x6f9dbe4c39175153, 0xf2fec757eb99e88, 0x43d7361a93733d91, 0x3abd10ed19c85a3d, 0xc4de9ab9c5ef7181, 0x53e375901684c900, 0x68ffc3e7d71c41ff, 0x47adab62c8d942fe, 0x226a33fd6fbb381d, 0x87ef4c8fdd83309a, 0xaca1cf44c5fa8799, 0x6cbae86c755f},
	B: Fp751Element{0x4c80c37fe68282a7, 0xbd8b9d7248bf553a, 0x1fb0e8e74d5e1762, 0xb63fa0e4e5f91482, 0xc675ab8a45a1439, 0xdfa6772deace7820, 0xf0d813d71d9a9255, 0x53a1a58c634534bd, 0x4ebfc6485fdfd888, 0x6991fe4358bcf169, 0xc0547bdaca85b6fd, 0xf461548d632}}

// x([3^2]P) = 3957171963425208493644602380039721164492341594850197356580248639045894821895524981729970650520936632013218950972842867220898274664982599375786979902471523505057611521217523103474682939638645404445093536997296151472632038973463*i + 1357869545269286021642168835877253886774707209614159162748874474269328421720121175566245719916322684751967981171882659798149072149161259103020057556362998810229937432814792024248155991141511691087135859252304684633946087474060
var affine_xP9 = ExtensionFieldElement{
	A: Fp751Element{0x7c0daa0f04ded4e0, 0x52dc4f883d85e065, 0x91afbdc2c1714d0b, 0xb7b3db8e658cfeba, 0x43d4e72a692882f3, 0x535c56d83753da30, 0xc8a58724433cbf5d, 0x351153c0a5e74219, 0x2c81827d19f93dd5, 0x26ef8aca3370ea1a, 0x1cf939a6dd225dec, 0x3403cb28ad41},
	B: Fp751Element{0x93e7bc373a9ff7b, 0x57b8cc47635ebc0f, 0x92eab55689106cf3, 0x93643111d421f24c, 0x1c58b519506f6b7a, 0xebd409fb998faa13, 0x5c86ed799d09d80e, 0xd9a1d764d6363562, 0xf95e87f92fb0c4cc, 0x6b2bbaf5632a5609, 0x2d9b6a809dfaff7f, 0x29c0460348b}}

// m = 96550223052359874398280314003345143371473380422728857598463622014420884224892
var mScalarBytes = [...]uint8{0x7c, 0x7b, 0x95, 0xfa, 0xb4, 0x75, 0x6c, 0x48, 0x8c, 0x17, 0x55, 0xb4, 0x49, 0xf5, 0x1e, 0xa3, 0xb, 0x31, 0xf0, 0xa4, 0xa6, 0x81, 0xad, 0x94, 0x51, 0x11, 0xe7, 0xf5, 0x5b, 0x7d, 0x75, 0xd5}

// x([m]P) = 7893578558852400052689739833699289348717964559651707250677393044951777272628231794999463214496545377542328262828965953246725804301238040891993859185944339366910592967840967752138115122568615081881937109746463885908097382992642*i + 8293895847098220389503562888233557012043261770526854885191188476280014204211818299871679993460086974249554528517413590157845430186202704783785316202196966198176323445986064452630594623103149383929503089342736311904030571524837
var affine_xaP = ExtensionFieldElement{
	A: Fp751Element{0x2112f3c7d7f938bb, 0x704a677f0a4df08f, 0x825370e31fb4ef00, 0xddbf79b7469f902, 0x27640c899ea739fd, 0xfb7b8b19f244108e, 0x546a6679dd3baebc, 0xe9f0ecf398d5265f, 0x223d2b350e75e461, 0x84b322a0b6aff016, 0xfabe426f539f8b39, 0x4507a0604f50},
	B: Fp751Element{0xac77737e5618a5fe, 0xf91c0e08c436ca52, 0xd124037bc323533c, 0xc9a772bf52c58b63, 0x3b30c8f38ef6af4d, 0xb9eed160e134f36e, 0x24e3836393b25017, 0xc828be1b11baf1d9, 0x7b7dab585df50e93, 0x1ca3852c618bd8e0, 0x4efa73bcb359fa00, 0x50b6a923c2d4}}

// Inputs for testing 3-point-ladder
var threePointLadderInputs = []ProjectivePoint{
	// x(P)
	ProjectivePoint{
		X: ExtensionFieldElement{
			A: Fp751Element{0xe8d05f30aac47247, 0x576ec00c55441de7, 0xbf1a8ec5fe558518, 0xd77cb17f77515881, 0x8e9852837ee73ec4, 0x8159634ad4f44a6b, 0x2e4eb5533a798c5, 0x9be8c4354d5bc849, 0xf47dc61806496b84, 0x25d0e130295120e0, 0xdbef54095f8139e3, 0x5a724f20862c},
			B: Fp751Element{0x3ca30d7623602e30, 0xfb281eddf45f07b7, 0xd2bf62d5901a45bc, 0xc67c9baf86306dd2, 0x4e2bd93093f538ca, 0xcfd92075c25b9cbe, 0xceafe9a3095bcbab, 0x7d928ad380c85414, 0x37c5f38b2afdc095, 0x75325899a7b779f4, 0xf130568249f20fdd, 0x178f264767d1}},
		Z: oneExtensionField,
	},
	// x(Q)
	ProjectivePoint{
		X: ExtensionFieldElement{
			A: Fp751Element{0x2b71a2a93ad1e10e, 0xf0b9842a92cfb333, 0xae17373615a27f5c, 0x3039239f428330c4, 0xa0c4b735ed7dcf98, 0x6e359771ddf6af6a, 0xe986e4cac4584651, 0x8233a2b622d5518, 0xbfd67bf5f06b818b, 0xdffe38d0f5b966a6, 0xa86b36a3272ee00a, 0x193e2ea4f68f},
			B: Fp751Element{0x5a0f396459d9d998, 0x479f42250b1b7dda, 0x4016b57e2a15bf75, 0xc59f915203fa3749, 0xd5f90257399cf8da, 0x1fb2dadfd86dcef4, 0x600f20e6429021dc, 0x17e347d380c57581, 0xc1b0d5fa8fe3e440, 0xbcf035330ac20e8, 0x50c2eb5f6a4f03e6, 0x86b7c4571}},
		Z: oneExtensionField,
	},
	// x(P-Q)
	ProjectivePoint{
		X: ExtensionFieldElement{
			A: Fp751Element{0x4aafa9f378f7b5ff, 0x1172a683aa8eee0, 0xea518d8cbec2c1de, 0xe191bcbb63674557, 0x97bc19637b259011, 0xdbeae5c9f4a2e454, 0x78f64d1b72a42f95, 0xe71cb4ea7e181e54, 0xe4169d4c48543994, 0x6198c2286a98730f, 0xd21d675bbab1afa5, 0x2e7269fce391},
			B: Fp751Element{0x23355783ce1d0450, 0x683164cf4ce3d93f, 0xae6d1c4d25970fd8, 0x7807007fb80b48cf, 0xa005a62ec2bbb8a2, 0x6b5649bd016004cb, 0xbb1a13fa1330176b, 0xbf38e51087660461, 0xe577fddc5dd7b930, 0x5f38116f56947cd3, 0x3124f30b98c36fde, 0x4ca9b6e6db37}},
		Z: oneExtensionField,
	},
}

// Helpers

func (P ProjectivePoint) Generate(rand *rand.Rand, size int) reflect.Value {
	f := ExtensionFieldElement{}
	x, _ := f.Generate(rand, size).Interface().(ExtensionFieldElement)
	z, _ := f.Generate(rand, size).Interface().(ExtensionFieldElement)
	return reflect.ValueOf(ProjectivePoint{
		X: x,
		Z: z,
	})
}

func (curve ProjectiveCurveParameters) Generate(rand *rand.Rand, size int) reflect.Value {
	f := ExtensionFieldElement{}
	A, _ := f.Generate(rand, size).Interface().(ExtensionFieldElement)
	C, _ := f.Generate(rand, size).Interface().(ExtensionFieldElement)
	return reflect.ValueOf(ProjectiveCurveParameters{
		A: A,
		C: C,
	})
}

// Sets FP(p^2) from uint64. x sets ExtensionsFieldElement.A (real part) and y
// ExtensionFieldElement.B (imaginary part).
//
// Returns dest to allow chaining operations.
func (dest *ExtensionFieldElement) SetUint64(x, y uint64) *ExtensionFieldElement {
	var xRR fp751X2
	dest.A = Fp751Element{}                 // = 0
	dest.A[0] = x                           // = x
	fp751Mul(&xRR, &dest.A, &montgomeryRsq) // = x*R*R
	fp751MontgomeryReduce(&dest.A, &xRR)    // = x*R mod p

	dest.B = Fp751Element{}                 // = 0
	dest.B[0] = y                           // = y
	fp751Mul(&xRR, &dest.B, &montgomeryRsq) // = y*R*R
	fp751MontgomeryReduce(&dest.B, &xRR)    // = y*R mod p

	return dest
}

// Helpers

// Given xP = x(P), xQ = x(Q), and xPmQ = x(P-Q), compute xR = x(P+Q).
//
// Returns xR to allow chaining.  Safe to overlap xP, xQ, xR.
func (xR *ProjectivePoint) Add(xP, xQ, xPmQ *ProjectivePoint) *ProjectivePoint {
	// Algorithm 1 of Costello-Smith.
	var v0, v1, v2, v3, v4 ExtensionFieldElement
	v0.Add(&xP.X, &xP.Z)               // X_P + Z_P
	v1.Sub(&xQ.X, &xQ.Z).Mul(&v1, &v0) // (X_Q - Z_Q)(X_P + Z_P)
	v0.Sub(&xP.X, &xP.Z)               // X_P - Z_P
	v2.Add(&xQ.X, &xQ.Z).Mul(&v2, &v0) // (X_Q + Z_Q)(X_P - Z_P)
	v3.Add(&v1, &v2).Square(&v3)       // 4(X_Q X_P - Z_Q Z_P)^2
	v4.Sub(&v1, &v2).Square(&v4)       // 4(X_Q Z_P - Z_Q X_P)^2
	v0.Mul(&xPmQ.Z, &v3)               // 4X_{P-Q}(X_Q X_P - Z_Q Z_P)^2
	xR.Z.Mul(&xPmQ.X, &v4)             // 4Z_{P-Q}(X_Q Z_P - Z_Q X_P)^2
	xR.X = v0
	return xR
}

// Given xP = x(P) and cached curve parameters Aplus2C = A + 2*C, C4 = 4*C,
// compute xQ = x([2]P).
//
// Returns xQ to allow chaining.  Safe to overlap xP, xQ.
func (xQ *ProjectivePoint) Double(xP *ProjectivePoint, Aplus2C, C4 *ExtensionFieldElement) *ProjectivePoint {
	// Algorithm 2 of Costello-Smith, amended to work with projective curve coefficients.
	var v1, v2, v3, xz4 ExtensionFieldElement
	v1.Add(&xP.X, &xP.Z).Square(&v1) // (X+Z)^2
	v2.Sub(&xP.X, &xP.Z).Square(&v2) // (X-Z)^2
	xz4.Sub(&v1, &v2)                // 4XZ = (X+Z)^2 - (X-Z)^2
	v2.Mul(&v2, C4)                  // 4C(X-Z)^2
	xQ.X.Mul(&v1, &v2)               // 4C(X+Z)^2(X-Z)^2
	v3.Mul(&xz4, Aplus2C)            // 4XZ(A + 2C)
	v3.Add(&v3, &v2)                 // 4XZ(A + 2C) + 4C(X-Z)^2
	xQ.Z.Mul(&v3, &xz4)              // (4XZ(A + 2C) + 4C(X-Z)^2)4XZ
	// Now (xQ.x : xQ.z)
	//   = (4C(X+Z)^2(X-Z)^2 : (4XZ(A + 2C) + 4C(X-Z)^2)4XZ )
	//   = ((X+Z)^2(X-Z)^2 : (4XZ((A + 2C)/4C) + (X-Z)^2)4XZ )
	//   = ((X+Z)^2(X-Z)^2 : (4XZ((a + 2)/4) + (X-Z)^2)4XZ )
	return xQ
}

// Given x(P) and a scalar m in little-endian bytes, compute x([m]P) using the
// Montgomery ladder.  This is described in Algorithm 8 of Costello-Smith.
//
// This function's execution time is dependent only on the byte-length of the
// input scalar.  All scalars of the same input length execute in uniform time.
// The scalar can be padded with zero bytes to ensure a uniform length.
//
// Safe to overlap the source with the destination.
func (xQ *ProjectivePoint) ScalarMult(curve *ProjectiveCurveParameters, xP *ProjectivePoint, scalar []uint8) *ProjectivePoint {
	var x0, x1, tmp ProjectivePoint
	var Aplus2C, C4 ExtensionFieldElement

	Aplus2C.Add(&curve.C, &curve.C) // = 2*C
	C4.Add(&Aplus2C, &Aplus2C)      // = 4*C
	Aplus2C.Add(&Aplus2C, &curve.A) // = 2*C + A

	x0.X.One()
	x0.Z.Zero()
	x1 = *xP

	// Iterate over the bits of the scalar, top to bottom
	prevBit := uint8(0)
	for i := len(scalar) - 1; i >= 0; i-- {
		scalarByte := scalar[i]
		for j := 7; j >= 0; j-- {
			bit := (scalarByte >> uint(j)) & 0x1
			ProjectivePointConditionalSwap(&x0, &x1, (bit ^ prevBit))
			tmp.Double(&x0, &Aplus2C, &C4)
			x1.Add(&x0, &x1, xP)
			x0 = tmp
			prevBit = bit
		}
	}
	// now prevBit is the lowest bit of the scalar
	ProjectivePointConditionalSwap(&x0, &x1, prevBit)
	*xQ = x0
	return xQ
}

// Tests

func TestOne(t *testing.T) {
	var tmp ExtensionFieldElement

	tmp.Mul(&oneExtensionField, &affine_xP)
	if !tmp.VartimeEq(&affine_xP) {
		t.Error("Not equal 1")
	}
}

func TestScalarMultVersusSage(t *testing.T) {
	var xP ProjectivePoint

	xP.FromAffine(&affine_xP)
	affine_xQ := xP.ScalarMult(&curve, &xP, mScalarBytes[:]).ToAffine() // = x([m]P)
	if !affine_xaP.VartimeEq(affine_xQ) {
		t.Error("\nExpected\n", affine_xaP, "\nfound\n", affine_xQ)
	}
}

func Test_jInvariant(t *testing.T) {
	var curve = ProjectiveCurveParameters{A: curve_A, C: curve_C}
	var jbufRes = make([]byte, P751_SharedSecretSize)
	var jbufExp = make([]byte, P751_SharedSecretSize)
	// Computed using Sage
	// j = 3674553797500778604587777859668542828244523188705960771798425843588160903687122861541242595678107095655647237100722594066610650373491179241544334443939077738732728884873568393760629500307797547379838602108296735640313894560419*i + 3127495302417548295242630557836520229396092255080675419212556702820583041296798857582303163183558315662015469648040494128968509467224910895884358424271180055990446576645240058960358037224785786494172548090318531038910933793845
	var known_j = ExtensionFieldElement{
		A: Fp751Element{0xc7a8921c1fb23993, 0xa20aea321327620b, 0xf1caa17ed9676fa8, 0x61b780e6b1a04037, 0x47784af4c24acc7a, 0x83926e2e300b9adf, 0xcd891d56fae5b66, 0x49b66985beb733bc, 0xd4bcd2a473d518f, 0xe242239991abe224, 0xa8af5b20f98672f8, 0x139e4d4e4d98},
		B: Fp751Element{0xb5b52a21f81f359, 0x715e3a865db6d920, 0x9bac2f9d8911978b, 0xef14acd8ac4c1e3d, 0xe81aacd90cfb09c8, 0xaf898288de4a09d9, 0xb85a7fb88c5c4601, 0x2c37c3f1dd303387, 0x7ad3277fe332367c, 0xd4cbee7f25a8e6f8, 0x36eacbe979eaeffa, 0x59eb5a13ac33},
	}

	curve.Jinvariant(jbufRes)
	known_j.ToBytes(jbufExp)

	if !bytes.Equal(jbufRes, jbufExp) {
		t.Error("Computed incorrect j-invariant: found\n", jbufRes, "\nexpected\n", jbufExp)
	}
}

func TestProjectivePointVartimeEq(t *testing.T) {
	var xP ProjectivePoint

	xP.FromAffine(&affine_xP)
	xQ := xP
	// Scale xQ, which results in the same projective point
	xQ.X.Mul(&xQ.X, &curve_A)
	xQ.Z.Mul(&xQ.Z, &curve_A)
	if !xQ.VartimeEq(&xP) {
		t.Error("Expected the scaled point to be equal to the original")
	}
}

func TestPointDoubleVersusSage(t *testing.T) {
	var curve = ProjectiveCurveParameters{A: curve_A, C: curve_C}
	var params = curve.CalcCurveParamsEquiv4()
	var xP, xQ ProjectivePoint

	xP.FromAffine(&affine_xP)
	affine_xQ := xQ.Pow2k(&params, &xP, 1).ToAffine()
	if !affine_xQ.VartimeEq(&affine_xP2) {
		t.Error("\nExpected\n", affine_xP2, "\nfound\n", affine_xQ)
	}
}

func TestPointMul4VersusSage(t *testing.T) {
	var params = curve.CalcCurveParamsEquiv4()
	var xP, xQ ProjectivePoint

	xP.FromAffine(&affine_xP)
	affine_xQ := xQ.Pow2k(&params, &xP, 2).ToAffine()
	if !affine_xQ.VartimeEq(&affine_xP4) {
		t.Error("\nExpected\n", affine_xP4, "\nfound\n", affine_xQ)
	}
}

func TestPointMul9VersusSage(t *testing.T) {
	var params = curve.CalcCurveParamsEquiv3()
	var xP, xQ ProjectivePoint

	xP.FromAffine(&affine_xP)
	affine_xQ := xQ.Pow3k(&params, &xP, 2).ToAffine()
	if !affine_xQ.VartimeEq(&affine_xP9) {
		t.Error("\nExpected\n", affine_xP9, "\nfound\n", affine_xQ)
	}
}

func TestPointPow2kVersusScalarMult(t *testing.T) {
	var xP, xQ, xR ProjectivePoint
	var params = curve.CalcCurveParamsEquiv4()

	xP.FromAffine(&affine_xP)
	affine_xQ := xQ.Pow2k(&params, &xP, 5).ToAffine()              // = x([32]P)
	affine_xR := xR.ScalarMult(&curve, &xP, []byte{32}).ToAffine() // = x([32]P)

	if !affine_xQ.VartimeEq(affine_xR) {
		t.Error("\nExpected\n", affine_xQ, "\nfound\n", affine_xR)
	}
}

func TestRecoverCoordinateA(t *testing.T) {
	var cparam ProjectiveCurveParameters
	// Vectors generated with SIKE reference implementation
	var a = ExtensionFieldElement{
		A: Fp751Element{0x9331D9C5AAF59EA4, 0xB32B702BE4046931, 0xCEBB333912ED4D34, 0x5628CE37CD29C7A2, 0x0BEAC5ED48B7F58E, 0x1FB9D3E281D65B07, 0x9C0CFACC1E195662, 0xAE4BCE0F6B70F7D9, 0x59E4E63D43FE71A0, 0xEF7CE57560CC8615, 0xE44A8FB7901E74E8, 0x000069D13C8366D1},
		B: Fp751Element{0xF6DA1070279AB966, 0xA78FB0CE7268C762, 0x19B40F044A57ABFA, 0x7AC8EE6160C0C233, 0x93D4993442947072, 0x757D2B3FA4E44860, 0x073A920F8C4D5257, 0x2031F1B054734037, 0xDEFAA1D2406555CD, 0x26F9C70E1496BE3D, 0x5B3F335A0A4D0976, 0x000013628B2E9C59}}
	var affine_xP = ExtensionFieldElement{
		A: Fp751Element{0xea6b2d1e2aebb250, 0x35d0b205dc4f6386, 0xb198e93cb1830b8d, 0x3b5b456b496ddcc6, 0x5be3f0d41132c260, 0xce5f188807516a00, 0x54f3e7469ea8866d, 0x33809ef47f36286, 0x6fa45f83eabe1edb, 0x1b3391ae5d19fd86, 0x1e66daf48584af3f, 0xb430c14aaa87},
		B: Fp751Element{0x97b41ebc61dcb2ad, 0x80ead31cb932f641, 0x40a940099948b642, 0x2a22fd16cdc7fe84, 0xaabf35b17579667f, 0x76c1d0139feb4032, 0x71467e1e7b1949be, 0x678ca8dadd0d6d81, 0x14445daea9064c66, 0x92d161eab4fa4691, 0x8dfbb01b6b238d36, 0x2e3718434e4e}}
	var affine_xQ = ExtensionFieldElement{
		A: Fp751Element{0xb055cf0ca1943439, 0xa9ff5de2fa6c69ed, 0x4f2761f934e5730a, 0x61a1dcaa1f94aa4b, 0xce3c8fadfd058543, 0xeac432aaa6701b8e, 0x8491d523093aea8b, 0xba273f9bd92b9b7f, 0xd8f59fd34439bb5a, 0xdc0350261c1fe600, 0x99375ab1eb151311, 0x14d175bbdbc5},
		B: Fp751Element{0xffb0ef8c2111a107, 0x55ceca3825991829, 0xdbf8a1ccc075d34b, 0xb8e9187bd85d8494, 0x670aa2d5c34a03b0, 0xef9fe2ed2b064953, 0xc911f5311d645aee, 0xf4411f409e410507, 0x934a0a852d03e1a8, 0xe6274e67ae1ad544, 0x9f4bc563c69a87bc, 0x6f316019681e}}
	var affine_xQmP = ExtensionFieldElement{
		A: Fp751Element{0x6ffb44306a153779, 0xc0ffef21f2f918f3, 0x196c46d35d77f778, 0x4a73f80452edcfe6, 0x9b00836bce61c67f, 0x387879418d84219e, 0x20700cf9fc1ec5d1, 0x1dfe2356ec64155e, 0xf8b9e33038256b1c, 0xd2aaf2e14bada0f0, 0xb33b226e79a4e313, 0x6be576fad4e5},
		B: Fp751Element{0x7db5dbc88e00de34, 0x75cc8cb9f8b6e11e, 0x8c8001c04ebc52ac, 0x67ef6c981a0b5a94, 0xc3654fbe73230738, 0xc6a46ee82983ceca, 0xed1aa61a27ef49f0, 0x17fe5a13b0858fe0, 0x9ae0ca945a4c6b3c, 0x234104a218ad8878, 0xa619627166104394, 0x556a01ff2e7e}}

	cparam.RecoverCoordinateA(&affine_xP, &affine_xQ, &affine_xQmP)
	cparam.C.One()

	// Check A is correct
	if !cparam.A.VartimeEq(&a) {
		t.Error("\nExpected\n", a, "\nfound\n", cparam.A)
	}

	// Check C is not changed
	if !cparam.C.VartimeEq(&oneExtensionField) {
		t.Error("\nExpected\n", cparam.C, "\nfound\n", oneExtensionField)
	}
}

func TestR2LVersusSage(t *testing.T) {
	var xR ProjectivePoint

	sageAffine_xR := ExtensionFieldElement{
		A: Fp751Element{0x729465ba800d4fd5, 0x9398015b59e514a1, 0x1a59dd6be76c748e, 0x1a7db94eb28dd55c, 0x444686e680b1b8ec, 0xcc3d4ace2a2454ff, 0x51d3dab4ec95a419, 0xc3b0f33594acac6a, 0x9598a74e7fd44f8a, 0x4fbf8c638f1c2e37, 0x844e347033052f51, 0x6cd6de3eafcf},
		B: Fp751Element{0x85da145412d73430, 0xd83c0e3b66eb3232, 0xd08ff2d453ec1369, 0xa64aaacfdb395b13, 0xe9cba211a20e806e, 0xa4f80b175d937cfc, 0x556ce5c64b1f7937, 0xb59b39ea2b3fdf7a, 0xc2526b869a4196b3, 0x8dad90bca9371750, 0xdfb4a30c9d9147a2, 0x346d2130629b}}
	xR = RightToLeftLadder(&curve, &threePointLadderInputs[0], &threePointLadderInputs[1], &threePointLadderInputs[2], uint(len(mScalarBytes)*8), mScalarBytes[:])
	affine_xR := xR.ToAffine()

	if !affine_xR.VartimeEq(&sageAffine_xR) {
		t.Error("\nExpected\n", sageAffine_xR, "\nfound\n", affine_xR)
	}
}

func TestPointTripleVersusAddDouble(t *testing.T) {
	tripleEqualsAddDouble := func(curve ProjectiveCurveParameters, P ProjectivePoint) bool {
		var P2, P3, P2plusP ProjectivePoint

		eqivParams4 := curve.CalcCurveParamsEquiv4()
		eqivParams3 := curve.CalcCurveParamsEquiv3()
		P2.Pow2k(&eqivParams4, &P, 1) // = x([2]P)
		P3.Pow3k(&eqivParams3, &P, 1) // = x([3]P)
		P2plusP.Add(&P2, &P, &P)      // = x([2]P + P)
		return P3.VartimeEq(&P2plusP)
	}

	if err := quick.Check(tripleEqualsAddDouble, quickCheckConfig); err != nil {
		t.Error(err)
	}
}

func TestProjectiveParamsRecovery(t *testing.T) {
	var crv1, crv2 ProjectiveCurveParameters
	var A4, C4, four ExtensionFieldElement
	four.SetUint64(4, 0)

	eqivParams3 := curve.CalcCurveParamsEquiv3()
	eqivParams4 := curve.CalcCurveParamsEquiv4()

	// This returns 4*A and 4*C
	crv1.RecoverCurveCoefficients3(&eqivParams3)
	crv2.RecoverCurveCoefficients4(&eqivParams4)

	A4.Mul(&curve_A, &four)
	C4.Mul(&curve_C, &four)

	if !crv1.A.VartimeEq(&A4) {
		t.Error("\nExpected\n", crv1.A, "\nfound\n", A4)
	}
	if !crv1.C.VartimeEq(&C4) {
		t.Error("\nExpected\n", crv1.C, "\nfound\n", C4)
	}

	if !crv2.A.VartimeEq(&curve_A) {
		t.Error("\nExpected\n", crv2.A, "\nfound\n", curve_A)
	}
	if !crv2.C.VartimeEq(&curve_C) {
		t.Error("\nExpected\n", crv2.C, "\nfound\n", curve_C)
	}
}

func BenchmarkThreePointLadder379BitScalar(b *testing.B) {
	var mScalarBytes = [...]uint8{84, 222, 146, 63, 85, 18, 173, 162, 167, 38, 10, 8, 143, 176, 93, 228, 247, 128, 50, 128, 205, 42, 15, 137, 119, 67, 43, 3, 61, 91, 237, 24, 235, 12, 53, 96, 186, 164, 232, 223, 197, 224, 64, 109, 137, 63, 246, 4}

	for n := 0; n < b.N; n++ {
		RightToLeftLadder(&curve, &threePointLadderInputs[0], &threePointLadderInputs[1], &threePointLadderInputs[2], uint(len(mScalarBytes)*8), mScalarBytes[:])
	}
}

func BenchmarkR2L379BitScalar(b *testing.B) {
	var mScalarBytes = [...]uint8{84, 222, 146, 63, 85, 18, 173, 162, 167, 38, 10, 8, 143, 176, 93, 228, 247, 128, 50, 128, 205, 42, 15, 137, 119, 67, 43, 3, 61, 91, 237, 24, 235, 12, 53, 96, 186, 164, 232, 223, 197, 224, 64, 109, 137, 63, 246, 4}

	for n := 0; n < b.N; n++ {
		RightToLeftLadder(&curve, &threePointLadderInputs[0], &threePointLadderInputs[1], &threePointLadderInputs[2], uint(len(mScalarBytes)*8), mScalarBytes[:])
	}
}
